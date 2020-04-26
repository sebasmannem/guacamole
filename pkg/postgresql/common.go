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
	"fmt"
	"go.uber.org/zap"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	walSegSize              = 16777216 // 16MiB
	maxConnections          = 1
	PgUnixSocketDirectories = "/tmp"
)

const (
	reNWHostTypes = `host|hostssl|hostnossql`
	reHostTypes   = `local|host|hostssl|hostnossql`
	reMethods     = "trust|reject|scram-sha-256|md5|password|gss|sspi|ident|peer|ldap|radius|cert|pam|bsd"
	reObject      = `[a-zA-Z0-9_]+`
	reIPAddress   = `[0-9:][0-9.:A-Fa-f]+`
	reIPNetmask   = `[0-9]{1,3}`
	reHostname    = `\.?[a-zA-Z][a-zA-Z0-9_.-]+`
)

var (
	validReplSlotName = regexp.MustCompile("^[a-z0-9_]+$")
	sugar             *zap.SugaredLogger
)

// Initialize function can be used to initialize the module
func Initialize(log *zap.SugaredLogger) {
	sugar = log
	sugar.Debugf("postgresql module is Initialized")
}

func replSlotNameValid(name string) bool {
	return validReplSlotName.MatchString(name)
}

func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func expand(s, dataDir string) string {
	buf := make([]byte, 0, 2*len(s))
	// %d %% are all ASCII, so bytes are fine for this operation.
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '%' && j+1 < len(s) {
			switch s[j+1] {
			case 'd':
				buf = append(buf, s[i:j]...)
				buf = append(buf, []byte(dataDir)...)
				j++
				i = j + 1
			case '%':
				j++
				buf = append(buf, s[i:j]...)
				i = j + 1
			default:
			}
		}
	}
	return string(buf) + s[i:]
}

// pgLsnToInt function can be used to calculate the exact byte address from a LSN
func pgLsnToInt(lsn string) (uint64, error) {
	parts := strings.Split(lsn, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("bad pg_lsn: %s", lsn)
	}
	a, err := strconv.ParseUint(parts[0], 16, 32)
	if err != nil {
		return 0, err
	}
	b, err := strconv.ParseUint(parts[1], 16, 32)
	if err != nil {
		return 0, err
	}
	v := uint64(a)<<32 | b
	return v, nil
}


func isWalFileName(name string) bool {
	walChars := "0123456789ABCDEF"
	if len(name) != 24 {
		return false
	}
	for _, c := range name {
		ok := false
		for _, v := range walChars {
			if c == v {
				ok = true
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

//func XlogPosTowalFileNameNoTimeline(XLogPos uint64) string {
//	id := uint32(XLogPos >> 32)
//	offset := uint32(XLogPos)
//	// TODO(sgotti) for now we assume wal size is the default 16M size
//	seg := offset / walSegSize
//	return fmt.Sprintf("%08X%08X", id, seg)
//}

func walFileNameNoTimeLine(name string) (string, error) {
	if !isWalFileName(name) {
		return "", fmt.Errorf("bad wal file name")
	}
	return name[8:24], nil
}

func reQuickTest(value string, re string) int {
	match, err := regexp.MatchString(fmt.Sprintf("^(%s)$", re), value)
	if err != nil {
		fmt.Printf("%s", err)
		return 1
	} else if !match {
		fmt.Printf("hba line field %s does not fit to re '^(%s)$'", value, re)
		return 1
	}
	return 0
}
