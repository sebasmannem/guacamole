package postgresql

import (
	"fmt"
	"regexp"
	"strings"
	"strconv"
)

// K8PGHbaKey type can store a hba rule. It can be retrieved from a configmap an/or a hba file
type (
	PgVersion struct {
		major int `json:"major"`
		minor int `json:"minor"`
	}
)

// NewK8PGHbaKey can be used to create a 8PGHbaEntry from the separate fields
func NewPgVersion(major int, minor int) (*PgVersion, error) {
	return &PgVersion{
		major: major,
		minor: minor,
	}, nil
}

// String can convert a K8PGHbaKey into an array of strings
func (version *PgVersion) String() string {
	return fmt.Sprintf("PgVersion{%s, %s}", version.major, version.minor)
}

// LoadFromBinary parses the output of a postgres binary with -V option, and reads the version information
func (version *PgVersion) LoadFromBinary(v string) error {
	captures := make(map[string]string)
	// extact version (removing beta*, rc* etc...)
	expr := `.*(?P<Database>\(PostgreSQL\)|\(EnterpriseDB\)) (?P<Version>[0-9\.]+).*`
	regex, err := regexp.Compile(expr)
	if err != nil {
		return err
	}
	m := regex.FindStringSubmatch(v)

	if len(m) != 3 {
		return fmt.Errorf("failed to parse postgres binary version: %q", v)
	}

	for i, name := range regex.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}

		captures[name] = m[i]

	}

	return version.parseVersion(captures["Version"])
}

// ParseVersion function
func (version *PgVersion) parseVersion(v string) error {
	v = string(strings.Split(v, "\n")[0])
	parts := strings.Split(v, ".")
	if len(parts) < 1 {
		return fmt.Errorf("bad version: %q", v)
	}
	maj, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("failed to parse major %q: %v", parts[0], err)
	}
	min := 0
	if len(parts) > 1 {
		min, err = strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("failed to parse minor %q: %v", parts[1], err)
		}
	}

	version.major = maj
	version.minor = min
	return nil
}
