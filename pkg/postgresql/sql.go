// Copyright 2019 Oscar M Herrera(KnowledgeSource Solutions Inc}
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied
// See the License for the specific language governing permissions and
// limitations under the License.

package postgresql

import (
	"context"
	"fmt"
	"github.com/lib/pq"
	"regexp"
	"strings"
)

const (
	roleOptions = `(SUPERUSER|NOSUPERUSER|CREATEDB|NOCREATEDB|CREATEROLE|NOCREATEROLE|INHERIT|NOINHERIT|LOGIN|NOLOGIN|REPLICATION|NOREPLICATION|BYPASSRLS|NOBYPASSRLS|CONNECTION +LIMIT +[-0-9]+|(ENCRYPTED +)?PASSWORD +'[^']*'|PASSWORD +NULL|VALID +UNTIL +'[^']*')`
)

type rowMap map[string]interface{}

func (p *Manager) runQuery(query string) error {
	err := p.Ping()
	if err != nil {
		return err
	}
	sugar.Debugf("Running query: %s", query)
	_, err = p.defaultConnection.Query(context.Background(), query)

	return err
}

func (p *Manager) runQueryGetRows(query string) ([]rowMap, error) {
	var results []rowMap
	err := p.Ping()
	if err != nil {
		return results, err
	}

	rows, err := p.defaultConnection.Query(context.Background(), query)
	if err != nil {
		return results, err
	}

	defer rows.Close()
	for rows.Next() {
		var v []interface{}
		var result rowMap
		v, err = rows.Values()
		if err != nil {
			sugar.Fatal(err)
		}

		for i := range v {
			result[string(rows.FieldDescriptions()[i].Name)] = v[i]
		}
		results = append(results, result)
	}

	return results, nil
}

func (p *Manager) runQueryGetOneValue(query string) (interface{}, error) {
	var result interface{}
	err := p.Ping()
	if err != nil {
		return result, err
	}

	rows, err := p.defaultConnection.Query(context.Background(), query)
	if err != nil {
		return result, err
	}

	defer rows.Close()
	for rows.Next() {
		var v []interface{}
		v, err = rows.Values()
		if err != nil {
			return nil, err
		}
		if len(v) == 0 {
			return nil, fmt.Errorf("Query %s has no returning columns", query)
		} else {
			return v[0], nil
		}
	}
	return nil, fmt.Errorf("Query %s has returned no rows", query)
}

func (p *Manager) CreateRole(username, password string, options []string, roles []string) error {
	quotedUserName := pq.QuoteIdentifier(username)
	err := p.runQuery(fmt.Sprintf(`create role %s;`, quotedUserName))
	if err != nil {
		sugar.Infof("Cannot create user %s (already exists). Will try to set options and grant roles as required.", username)
	}
	return p.AlterRole(username, password, options, roles)
}

func (p *Manager) AlterRole(username, password string, options []string, roles []string) error {
	quotedUserName := pq.QuoteIdentifier(username)
	quotedPassword := "NULL"
	if password != "" {
		quotedPassword = pq.QuoteLiteral(password)
	}
	err := p.runQuery(fmt.Sprintf(`alter role %s with password %s;`, quotedUserName, quotedPassword))
	if err != nil {
		return err
	}
	re := regexp.MustCompile(roleOptions)
	for _, option := range options {
		if !re.MatchString(option) {
			return fmt.Errorf("error: Option %s (for role %s) seems like an invalid option.", username, option)
		}
		err := p.runQuery(fmt.Sprintf(`alter role %s with %s;`, quotedUserName, option))
		if err != nil {
			return err
		}
	}
	var quotedRoleName string
	for _, role := range roles {
		quotedRoleName = pq.QuoteIdentifier(role)
		err := p.runQuery(fmt.Sprintf(`grant %s to %s;`, quotedRoleName, quotedUserName))
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Manager) getReplicationSlots(maj int) ([]string, error) {
	var query string
	replSlots := []string{}

	version, err := p.QueryVersion()
	if err != nil {
		return replSlots, err
	}
	if version.major < 10 {
		query = "select slot_name from pg_replication_slots;"
	} else {
		query = "select slot_name from pg_replication_slots where temporary is false;"
	}

	results, err := p.runQueryGetRows(query)
	if err != nil {
		return replSlots, err
	}

	for _, result := range results {
		slotname := result["slot_name"].(string)
		replSlots = append(replSlots, slotname)
	}

	return replSlots, nil
}

// CreateReplicationSlot method cam be used to create a replication slot
func (p *Manager) CreateReplicationSlot(slotname string) error {
	quotedSlotname := pq.QuoteLiteral(slotname)
	return p.runQuery(fmt.Sprintf("select pg_create_physical_replication_slot(%s)", quotedSlotname))
}

// DropReplicationSlot method cam be used to drop a replication slot
func (p *Manager) DropReplicationSlot(slotname string) error {
	quotedSlotname := pq.QuoteLiteral(slotname)
	return p.runQuery(fmt.Sprintf("select pg_drop_replication_slot(%s)", quotedSlotname))
}

// SyncStandbys function can be used to retrieve the sync standbys
func (p *Manager) SyncStandbys() ([]string, error) {
	syncStandbys := []string{}

	results, err := p.runQueryGetRows("select application_name, sync_state from pg_stat_replication")
	if err != nil {
		return syncStandbys, err
	}

	for _, result := range results {
		if result["sync_state"] == "sync" {
			standby := result["application_name"].(string)
			syncStandbys = append(syncStandbys, standby)
		}
	}
	return syncStandbys, nil
}

//
func (p *Manager) readSettingsFromPg() (Parameters, error) {
	var pgParameters = Parameters{}
	pfsExists, err := p.runQueryGetOneValue("select count(*) from information_schema.tables where table_schema = 'pg_catalog' and table_name = 'pg_file_settings'")
	if err != nil {
		return nil, err
	}

	if pfsExists.(int) > 0 {
		// NOTE If some pg_parameters that cannot be changed without a restart
		// are removed from the postgresql.conf file the view will contain some
		// rows with null name and setting and the error field set to the cause.
		// So we have to filter out these or the Scan will fail.
		rows, err := p.runQueryGetRows("select name, setting from pg_file_settings where name IS NOT NULL and setting IS NOT NULL")
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			name := row["name"].(string)
			value := row["setting"].(string)
			pgParameters[name] = value
		}
		return pgParameters, nil
	}

	// Fallback to pg_settings
	rows, err := p.runQueryGetRows("select name, setting, source from pg_settings")
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		name := row["name"].(string)
		value := row["setting"].(string)
		source := row["source"].(string)
		if source == "configuration file" {
			pgParameters[name] = value
		}
	}
	return pgParameters, nil
}

func (p *Manager) isRestartRequiredUsingPendingRestart() (bool, error) {
	numPendingRestart, err := p.runQueryGetOneValue("select count(*) from pg_settings where pending_restart;")
	if err != nil {
		return false, err
	}
	return (numPendingRestart.(int) > 0), nil
}

func (p *Manager) isRestartRequiredUsingPgSettingsContext(changedParams []string) (bool, error) {
	var escapedChangedParams []string
	// Lets first escape all params properly
	for _, param := range changedParams {
		escapedChangedParams = append(escapedChangedParams, pq.QuoteLiteral(param))
	}
	filter := strings.Join(escapedChangedParams, ",")
	result, err := p.runQueryGetOneValue(fmt.Sprintf("select count(*) from pg_settings where context = 'postmaster' and name = ANY(%s)", filter))
	if err != nil {
		return false, err
	}
	return (result.(int) > 0), nil
}

// PgVersion can be used to get the version information from a postgres connection
func (p *Manager) QueryVersion() (*PgVersion, error) {
	result, err := p.runQueryGetOneValue("show server_version;")
	if err != nil {
		return nil, err
	}
	version, err := NewPgVersion(0, 0)
	if err != nil {
		return nil, err
	}
	err = version.parseVersion(result.(string))
	if err != nil {
		return nil, err
	}
	return version, nil
}
