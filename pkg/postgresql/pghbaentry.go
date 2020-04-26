package postgresql

import (
	"fmt"
	"regexp"
	"strings"
)

// K8PGHbaEntry type can store a hba rule. It can be retrieved from a configmap an/or a hba file
type (
	K8PGHbaEntry struct {
		Key     *K8PGHbaKey
		Method  string `json:"method"`
		Options string `json:"options"`
		Comment string `json:"comment"`
	}
)

// NewK8PGHbaEntry can be used to create a 8PGHbaEntry from the separate fields
func NewK8PGHbaEntry(hostType string, database string, user string, address string, mask string, method string, options string, comment string) (*K8PGHbaEntry, error) {
	if hostType == "" {
		return &K8PGHbaEntry{
			Key:     nil,
			Method:  "",
			Options: "",
			Comment: comment,
		}, nil
	}
	key, err := NewK8PGHbaKey(hostType, database, user, address, mask)
	if err != nil {
		return nil, err
	}

	if reQuickTest(method, reMethods) > 0 {
		return nil, fmt.Errorf("hba line has an invalid method: %s", method)
	}

	return &K8PGHbaEntry{
		Key:     key,
		Method:  method,
		Options: options,
		Comment: comment,
	}, nil
}

// NewK8PGHbaEntry can be used to create a 8PGHbaEntry from a hba line
func NewK8PGHbaEntryFromString(hbaLine string) (*K8PGHbaEntry, error) {
	var hbaParts, comments string
	captures := make(map[string]string)
	captures["mask"] = ""
	if strings.Contains(hbaLine, "#") {
		parts := strings.SplitN(hbaLine, "#", 2)
		hbaParts, comments = parts[0], parts[1]
	} else {
		hbaParts, comments = hbaLine, ""
	}
	if strings.TrimLeft(hbaParts, " \t") == "" {
		return NewK8PGHbaEntry("", "", "", "", "", "", "", comments)
	}

	var replacer = strings.NewReplacer("\t", "", " ", "")

	parseRegEx1 := fmt.Sprintf(`^(?P<host>local)[\t| ]+(?P<database>%s)[\t| ]+(?P<user>%[1]s)[\t| ]+(?P<method>%s)[\t| ]*(?P<options>.*)`, reObject, reMethods)
	parseRegEx2 := fmt.Sprintf(`^(?P<host>%s)[\t| ]+(?P<database>%s)[\t| ]+(?P<user>%[2]s)[\t| ]+(?P<address>%s/%s)[\t| ]+(?P<method>%s)[\t| ]*(?P<options>.*)`, reHostTypes, reObject, reIPAddress, reIPNetmask, reMethods)
	parseRegEx3 := fmt.Sprintf(`^(?P<host>%s)[\t| ]+(?P<database>%s)[\t| ]+(?P<user>%[2]s)[\t| ]+(?P<address>%s)[\t| ]+(?P<mask>%[3]s)[\t| ]+(?P<method>%s)[\t| ]*(?P<options>.*)`, reHostTypes, reObject, reIPAddress, reMethods)
	parseRegEx4 := fmt.Sprintf(`^(?P<host>%s)[\t| ]+(?P<database>%s)[\t| ]+(?P<user>%[2]s)[\t| ]+(?P<address>%s)[\t| ]+(?P<method>%s)[\t| ]*(?P<options>.*)`, reHostTypes, reObject, reHostname, reMethods)

	parseRegExes := []string{parseRegEx1, parseRegEx2, parseRegEx3, parseRegEx4}
	for _, parseRegEx := range parseRegExes {
		regex, err := regexp.Compile(parseRegEx)
		if err != nil {
			return nil, err
		}
		m := regex.FindStringSubmatch(hbaParts)

		if len(m) == 0 {
			continue
		}

		for i, name := range regex.SubexpNames() {
			// Ignore the whole regexp match and unnamed groups
			if i == 0 || name == "" {
				continue
			}
			captures[name] = replacer.Replace(m[i])
		}

		return NewK8PGHbaEntry(captures["host"], captures["database"], captures["user"], captures["address"], captures["mask"], captures["method"], captures["options"], comments)
	}
	err := fmt.Errorf("hba line %s did not match any known hba line format", hbaLine)
	return nil, err
}

// String can convert a K8PGHbaEntry into an array of strings
func (hbaEntry *K8PGHbaEntry) IsComment() bool {
	if hbaEntry.Key == nil {
		return true
	}
	return false
}

// String can convert a K8PGHbaEntry into an array of strings
func (hbaEntry *K8PGHbaEntry) HasComment() bool {
	if hbaEntry.Comment == "" {
		return false
	}
	return true
}

// String can convert a K8PGHbaEntry into an array of strings
func (hbaEntry *K8PGHbaEntry) String() string {
	var comments string
	if hbaEntry.IsComment() {
		return fmt.Sprintf("#%s\n", hbaEntry.Comment)
	}
	if hbaEntry.HasComment() {
		comments = "\t\t#" + hbaEntry.Comment
	}
	keystring := hbaEntry.Key.String()
	return fmt.Sprintf("%s\t\t%s\t\t%s%s\n", keystring, hbaEntry.Method, hbaEntry.Options, comments)
}

// String can convert a K8PGHbaEntry into an array of strings
func (hbaEntry *K8PGHbaEntry) Equals(other K8PGHbaEntry) bool {
	if hbaEntry.Key != other.Key {
		return false
	}
	if hbaEntry.Method != other.Method {
		return false
	}
	if hbaEntry.Options != other.Options {
		return false
	}
	if hbaEntry.Comment != other.Comment {
		return false
	}
	return true
}
