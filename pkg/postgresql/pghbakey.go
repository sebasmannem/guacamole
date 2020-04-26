package postgresql

import (
	"fmt"
)

// K8PGHbaKey type can store a hba rule. It can be retrieved from a configmap an/or a hba file
type (
	K8PGHbaKey struct {
		Type     string `json:"type"`
		Database string `json:"database"`
		User     string `json:"user"`
		Address  string `json:"address"`
		Mask     string `json:"mask"`
	}
)

// NewK8PGHbaKey can be used to create a 8PGHbaEntry from the separate fields
func NewK8PGHbaKey(hostType string, database string, user string, address string, mask string) (*K8PGHbaKey, error) {
	var numErrors = 0
	numErrors += reQuickTest(hostType, reHostTypes)
	numErrors += reQuickTest(database, reObject)
	numErrors += reQuickTest(user, reObject)
	if hostType != "local" {
		reAddress := fmt.Sprintf("(%s(/%s)?|%s)", reIPAddress, reIPNetmask, reHostname)
		numErrors += reQuickTest(address, reAddress)
	}
	if mask != "" {
		numErrors += reQuickTest(mask, reIPAddress)
	}

	if numErrors > 0 {
		return nil, fmt.Errorf("hba key has one or more invalid fields")
	}

	return &K8PGHbaKey{
		Type:     hostType,
		Database: database,
		User:     user,
		Address:  address,
		Mask:     mask,
	}, nil
}

// String can convert a K8PGHbaKey into an array of strings
func (hbakey *K8PGHbaKey) String() string {
	return fmt.Sprintf("%s\t\t%s\t\t%s\t\t%s\t\t%s", hbakey.Type, hbakey.Database, hbakey.User, hbakey.Address, hbakey.Mask)
}
