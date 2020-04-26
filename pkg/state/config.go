package state

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

// Config will hold all config of this cluster. It is read from a ConfigMap
type (
	Config struct {
		BinDirectoryPtr  string
		DataDirectoryPtr string
		Locale           string
		Encoding         string
		DataChecksums    bool
		LocalPGHba       string
		PgPort           int
		SuDBConnection   string
		ReplConnection   string
		IsEDB            bool
		AuthMethod       string
		ReplAuthMethod   string
		LogLevel         string
		MaxConnections   int
		Database         string
	}
)

// NewConfig function to create a new (initialized) Config
func NewConfig() *Config {
	return &Config{
		DataDirectoryPtr: "/var/lib/pgsql",
		Locale:           "en_US.UTF8",
		Encoding:         "UTF8",
		DataChecksums:    true,
		LocalPGHba:       "./pg_hba.conf",
		PgPort:           5432,
		SuDBConnection:   "",
		ReplConnection:   "",
		IsEDB:            false,
		AuthMethod:       "trust",
		ReplAuthMethod:   "trust",
		LogLevel:         "DEBUG",
		MaxConnections:   10,
		Database:         "postgres",
	}
}

// LoadFromHash method can load a new Clusterdata from annotations
func (c *Config) LoadFromHash(hash map[string]string) {
	mapstructure.Decode(hash, c)
}

// SafeToHash method can save Clusterdata to annotations
func (c *Config) SafeToHash() (map[string]string, error) {
	values := map[string]string{}
	iVal := reflect.ValueOf(c).Elem()
	typ := iVal.Type()
	for i := 0; i < iVal.NumField(); i++ {
		f := iVal.Field(i)
		// You ca use tags here...
		// tag := typ.Field(i).Tag.Get("tagname")
		// Convert each type into a string for the url.Values string map
		var v string
		switch f.Interface().(type) {
		case int, int8, int16, int32, int64:
			v = strconv.FormatInt(f.Int(), 10)
		case uint, uint8, uint16, uint32, uint64:
			v = strconv.FormatUint(f.Uint(), 10)
		case float32:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 32)
		case float64:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 64)
		case []byte:
			v = string(f.Bytes())
		case string:
			v = f.String()
		default:
			json, err := json.Marshal(f)
			if err != nil {
				return map[string]string{}, err
			}
			v = string(json)
		}
		values[typ.Field(i).Name] = v
	}
	delete(values, "suDBConnection")
	delete(values, "replConnection")
	return values, nil
}
