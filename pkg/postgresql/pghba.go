package postgresql

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/sebasmannem/k8pgquay/pkg/common"
)

// K8PGHba type is an array of K8PGHbaEntryss
type K8PGHba struct {
	Entries  []K8PGHbaEntry `json:"hba"`
	dirty    bool           `json:"dirty"`
	filepath string         `json:"filepath"`
}

// LoadFromStringArray can be used to load a K8PGHba from an array of strings
func (k8pghba *K8PGHba) LoadFromStringArray(hba []string) error {
	var hbaEntries = []K8PGHbaEntry{}
	for _, hbaLine := range hba {
		entry, err := NewK8PGHbaEntryFromString(hbaLine)
		if err != nil {
			sugar.Errorf("Unable to parse line: %v", err)
			return err
		}
		k8pghba.Add(-1, *entry)
		hbaEntries = append(hbaEntries, *entry)
	}
	k8pghba.Entries = hbaEntries
	k8pghba.dirty = true
	return nil
}

// ConvertToStringArray can convert a K8PGHbaEntry into an array of strings
func (k8pghba *K8PGHba) ConvertToStringArray() []string {
	var hbaArray []string

	for _, hbaEntry := range k8pghba.Entries {
		hbaArray = append(hbaArray, hbaEntry.String())
	}
	return hbaArray
}

// ConvertHbaToJK8PGHba can be used to create a K8PGHba from an array of strings
func (k8pghba *K8PGHba) Add(index int, newEntry K8PGHbaEntry) error {
	/*This function will add entry to array:
	  * At index if index reached first (insert)
	  * At location of entry with same key if found first (replace)
		* At end if index was set too high (append)
		Note that comment lines are not replaced, but rather inserted
	*/
	var newEntries []K8PGHbaEntry
	var found bool = false
	for i, entry := range k8pghba.Entries {
		if i == index && !found {
			// not yet found, and at index, so insert item
			newEntries = append(newEntries, newEntry)
			k8pghba.dirty = true
			found = true
			// But we need to check the entry for equalness
		}
		if entry.IsComment() {
			// Comment lines are not replaced
			newEntries = append(newEntries, entry)
			found = true
			continue
		}
		if newEntry.Equals(entry) {
			newEntries = append(newEntries, entry)
			found = true
			continue
		}
		if newEntry.Key == entry.Key {
			if !found {
				// entry has same key, is no comment, and item was not found yet. replace
				newEntries = append(newEntries, newEntry)
				k8pghba.dirty = true
				found = true
			}
			continue
		}
		// Entry is not same, insert original entry
		newEntries = append(newEntries, entry)
	}
	if !found {
		// item was not found yet. append at end
		newEntries = append(newEntries, newEntry)
	}

	k8pghba.Entries = newEntries
	return nil
}

// ReadFromFile method can be used to read recovery config to this programs cache from the config files
func (k8pghba *K8PGHba) ReadFromFile(pgHbaPath string) error {
	if pgHbaPath == "" {
		pgHbaPath = k8pghba.filepath
	} else {
		k8pghba.filepath = pgHbaPath
	}

	sugar.Debugf("Reading %s", pgHbaPath)
	pgHbaContent, err := ioutil.ReadFile(pgHbaPath)
	if err != nil {
		sugar.Errorf("unable to read pg_hba.conf: Error: %v", err)
		return err
	}
	hba := strings.Split(string(pgHbaContent), "\n")
	err = k8pghba.LoadFromStringArray(hba)
	if err == nil {
		k8pghba.filepath = pgHbaPath
	}
	k8pghba.dirty = false
	return nil
}

// WriteToFile method can be used to write hba config from this programs cache to the config files
func (k8pghba *K8PGHba) WriteToFile(pgHbaPath string) (bool, error) {
	if !k8pghba.dirty {
		return false, nil
	}

	if pgHbaPath == "" {
		pgHbaPath = k8pghba.filepath
	}
	sugar.Debugf("Writing %s", pgHbaPath)
	err := common.WriteFileAtomicFunc(pgHbaPath, 0600,
		func(f io.Writer) error {
			for _, line := range k8pghba.Entries {
				if _, err := f.Write([]byte(line.String())); err != nil {
					return err
				}
			}
			sugar.Debugf("Done writing %s", pgHbaPath)
			return nil
		},
	)
	if err == nil {
		k8pghba.dirty = false
		k8pghba.filepath = pgHbaPath
		return true, nil
	}
	return false, err
}
