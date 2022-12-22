package defaults

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucassabreu/clockify-cli/pkg/cmd/time-entry/util/defaults"
	"github.com/lucassabreu/clockify-cli/strhlp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// DefaultTimeEntry has the default properties for the working directory
type DefaultTimeEntry struct {
	Workspace   string   `json:"workspace,omitempty"   yaml:"workspace,omitempty"`
	ProjectID   string   `json:"project,omitempty"     yaml:"project,omitempty"`
	TaskID      string   `json:"task,omitempty"        yaml:"task,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Billable    *bool    `json:"billable,omitempty"    yaml:"billable,omitempty"`
	TagIDs      []string `json:"tags,omitempty"        yaml:"tags,omitempty,flow"`
}

// ScanParam sets how ScanForDefaults should look for defaults
type ScanParam struct {
	Dir      string
	Filename string
}

// TimeEntryDefaults is a manager for the default time entry parameters on a
// folder
type TimeEntryDefaults interface {
	// Read scan the directory informed and its parents for the defaults
	// file
	Read() (DefaultTimeEntry, error)
	// Write persists the default values to the folder
	Write(DefaultTimeEntry) error
}

// NewTimeEntryDefaults creates a new instance of TimeEntryDefaults
func NewTimeEntryDefaults(p ScanParam) TimeEntryDefaults {
	return nil
}

type timeEntryDefaults struct {
	ScanParam
	DefaultTimeEntry
}

// Read scan the directory informed and its parents for the defaults
// file
func (t *timeEntryDefaults) Read() (defaults.DefaultTimeEntry, error) {
	panic("not implemented") // TODO: Implement
}

// Write persists the default values to the folder
func (t *timeEntryDefaults) Write(_ defaults.DefaultTimeEntry) error {
	n := filepath.Join(t.Dir, t.Filename)
	f, err := os.OpenFile(n, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}

	if strings.HasSuffix(f.Name(), "json") {
		return json.NewEncoder(f).Encode(d)
	}

	return yaml.NewEncoder(f).Encode(d)
}

func WriteDefaults(dir, filename string, d DefaultTimeEntry) error {
}

// ScanError wraps errors from scanning for the defaults file
type ScanError struct {
	Err error
}

// Error shows error message
func (s *ScanError) Error() string {
	return s.Unwrap().Error()
}

// Unwrap gives access to the error chain
func (s *ScanError) Unwrap() error {
	return s.Err
}

// DefaultsFileNotFoundErr is returned when the scan can't find any files
var DefaultsFileNotFoundErr = errors.New("defaults file not found")

const DEFAULT_FILENAME = ".clockify-defaults"

func ScanForDefaults(p ScanParam) func() (DefaultTimeEntry, error) {
	return func() (DefaultTimeEntry, error) {
		if p.Filename == "" {
			p.Filename = DEFAULT_FILENAME
		}

		dir := filepath.FromSlash(p.Dir)
		d := DefaultTimeEntry{}
		for {
			f, err := firstMatch(dir, p.Filename)
			if err != nil {
				return d, &ScanError{
					Err: errors.Wrap(
						err, "failed to open defaults file"),
				}
			}

			if f == nil {
				nDir := filepath.Dir(dir)
				if nDir == dir {
					return d, DefaultsFileNotFoundErr
				}

				dir = nDir
				continue
			}

			if strings.HasSuffix(f.Name(), "json") {
				err = json.NewDecoder(f).Decode(&d)
			} else {
				err = yaml.NewDecoder(f).Decode(&d)
			}

			if err != nil {
				return d, &ScanError{
					Err: errors.Wrap(
						err, "failed to decode defaults file"),
				}
			}

			return d, nil
		}
	}
}

func firstMatch(dir, filename string) (*os.File, error) {
	ms, _ := filepath.Glob(filepath.Join(dir, filename+".*"))
	if len(ms) == 0 {
		return nil, nil
	}

	ms = strhlp.Filter(
		func(s string) bool {
			return strings.HasSuffix(s, ".json") ||
				strings.HasSuffix(s, ".yml") ||
				strings.HasSuffix(s, ".yaml")
		},
		ms,
	)

	for _, m := range ms {
		entry, err := os.Open(m)
		if err != nil {
			return nil, err
		}

		s, err := entry.Stat()
		if err != nil || s.IsDir() {
			continue

		}
		return entry, nil
	}
	return nil, nil
}
