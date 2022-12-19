package defaults_test

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lucassabreu/clockify-cli/pkg/cmd/time-entry/util/defaults"
	"github.com/stretchr/testify/assert"
)

func TestWriteDefaults(t *testing.T) {
	tts := []struct {
		filename string
		d        defaults.DefaultTimeEntry
	}{
		{
			filename: "y_empty.yml",
			d:        defaults.DefaultTimeEntry{},
		},
		{
			filename: "j_empty.json",
			d:        defaults.DefaultTimeEntry{},
		},
		{
			filename: "j_complete.json",
			d: defaults.DefaultTimeEntry{
				Workspace:   "w",
				ProjectID:   "p",
				TaskID:      "t",
				Description: "d",
				TagIDs:      []string{"t1", "t2"},
			},
		},
		{
			filename: "y_complete.yaml",
			d: defaults.DefaultTimeEntry{
				Workspace:   "w",
				ProjectID:   "p",
				TaskID:      "t",
				Description: "d",
				TagIDs:      []string{"t1", "t2"},
			},
		},
	}

	dir := t.TempDir()
	for i := range tts {
		tt := &tts[i]
		t.Run(tt.filename, func(t *testing.T) {
			timeout(t, 6*time.Second, func() {
				err := defaults.WriteDefaults(dir, tt.filename, tt.d)
				if !assert.NoError(t, err) {
					return
				}

				f := strings.Split(tt.filename, ".")[0]

				r, err := defaults.ScanForDefaults(defaults.ScanParam{
					Dir:      dir,
					Filename: f,
				})()

				assert.NoError(t, err)
				assert.Equal(t, tt.d, r)
			})
		})
	}
}

func TestWriteDefaults_ShouldFail_WhenPermAreMissing(t *testing.T) {
	dir := t.TempDir()
	_ = os.Chmod(dir, 0444)
	timeout(t, 5*time.Second, func() {
		err := defaults.WriteDefaults(dir, "fail", defaults.DefaultTimeEntry{})
		assert.Error(t, err)
	})
}

func timeout(t *testing.T, d time.Duration, f func()) {
	done := make(chan bool)
	defer close(done)

	go func() {
		f()
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(d):
		t.Error("timeout " + d.String())
	}
}

func TestScanForDefaults_ShouldFail(t *testing.T) {
	wd, _ := os.Getwd()

	dir := t.TempDir()
	f, _ := os.OpenFile(
		filepath.Join(dir, "not-open.yaml"), os.O_CREATE, os.ModePerm)
	_ = f.Chmod(0000)
	_ = f.Close()

	tts := []struct {
		dir      string
		filename string
		err      interface{}
	}{
		{
			dir:      wd,
			filename: "not-found",
			err:      defaults.DefaultsFileNotFoundErr,
		},
		{
			dir:      filepath.Join(wd, "test_data", "test_cur"),
			filename: "not-right",
			err:      "invalid character",
		},
		{
			dir:      dir,
			filename: "not-open",
			err:      "permission denied",
		},
		{
			dir:      filepath.Join(wd, "test_data", "test_empty", "dir.yaml"),
			filename: "dir",
			err:      defaults.DefaultsFileNotFoundErr,
		},
	}

	for i := range tts {
		tt := &tts[i]
		t.Run(tt.filename, func(t *testing.T) {
			timeout(t, 5*time.Second, func() {
				d, err := defaults.ScanForDefaults(defaults.ScanParam{
					Dir:      tt.dir,
					Filename: tt.filename,
				})()

				assert.Equal(t, d, defaults.DefaultTimeEntry{})
				assert.Error(t, err)
				switch v := tt.err.(type) {
				case error:
					assert.ErrorIs(t, err, v)
				case string:
					assert.Regexp(t, v, err)
				}
			})
		})
	}
}

func TestScanForDefaults_ShouldLookUpperDirs(t *testing.T) {
	wd, _ := os.Getwd()
	tts := []struct {
		name     string
		param    defaults.ScanParam
		expected defaults.DefaultTimeEntry
	}{
		{
			name: "test_cur",
			param: defaults.ScanParam{
				Dir: "./test_data/test_cur",
			},
			expected: defaults.DefaultTimeEntry{
				Workspace:   "w",
				ProjectID:   "p",
				TaskID:      "t",
				Description: "d",
				TagIDs:      []string{"t1", "t2"},
			},
		},
		{
			name: "test_cur, filename as defaults",
			param: defaults.ScanParam{
				Dir:      "./test_data/test_cur",
				Filename: "defaults",
			},
			expected: defaults.DefaultTimeEntry{
				Workspace:   "W",
				ProjectID:   "P",
				TaskID:      "T",
				Description: "D",
			},
		},
		{
			name: "down again",
			param: defaults.ScanParam{
				Dir: "./test_data/test_cur/down/again",
			},
			expected: defaults.DefaultTimeEntry{
				Workspace:   "w",
				ProjectID:   "p",
				TaskID:      "t",
				Description: "d",
				TagIDs:      []string{"t1", "t2"},
			},
		},
		{
			name: "down path, filename as defaults",
			param: defaults.ScanParam{
				Dir:      "./test_data/test_cur/down/again",
				Filename: "defaults",
			},
			expected: defaults.DefaultTimeEntry{
				Workspace:   "W",
				ProjectID:   "P",
				TaskID:      "T",
				Description: "D",
			},
		},
		{
			name: "test_incompl",
			param: defaults.ScanParam{
				Dir: "./test_data/test_incompl",
			},
			expected: defaults.DefaultTimeEntry{
				Workspace: "w",
				ProjectID: "p",
			},
		},
		{
			name: "test_empty",
			param: defaults.ScanParam{
				Dir: "./test_data/test_empty/down/here",
			},
			expected: defaults.DefaultTimeEntry{},
		},
	}

	for i := range tts {
		tt := &tts[i]
		t.Run(tt.name, func(t *testing.T) {
			timeout(t, 1*time.Second, func() {
				tt.param.Dir = path.Join(wd, tt.param.Dir)
				d, _ := defaults.ScanForDefaults(tt.param)()
				assert.Equal(t, tt.expected, d)
			})
		})
	}
}
