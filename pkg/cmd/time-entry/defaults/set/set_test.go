package set_test

import (
	"io"
	"testing"

	"github.com/lucassabreu/clockify-cli/internal/mocks"
	"github.com/lucassabreu/clockify-cli/pkg/cmd/time-entry/defaults/set"
	"github.com/lucassabreu/clockify-cli/pkg/cmd/time-entry/util/defaults"
	"github.com/lucassabreu/clockify-cli/pkg/cmdutil"
	. "github.com/lucassabreu/clockify-cli/pkg/output/defaults"
	"github.com/stretchr/testify/assert"
)

var bTrue = true
var bFalse = false

func runCmd(f cmdutil.Factory, args []string) (
	d defaults.DefaultTimeEntry, reported bool, err error) {

	cmd := set.NewCmdSet(f, func(_ OutputFlags, _ io.Writer,
		dte defaults.DefaultTimeEntry) error {
		reported = true
		d = dte
		return nil
	})

	cmd.SilenceUsage = true
	cmd.SetArgs(args)
	_, err = cmd.ExecuteC()

	return d, reported, err
}

func TestNewCmdSet_ShouldFail_WhenInvalidArgs(t *testing.T) {
	tts := []struct {
		name    string
		args    []string
		err     string
		factory func(t *testing.T) cmdutil.Factory
	}{}

	for i := range tts {
		tt := &tts[i]
		t.Run(tt.name, func(t *testing.T) {
			f := mocks.NewMockFactory(t)
			f.EXPECT().Config().Return(&mocks.SimpleConfig{
				AllowNameForID: false,
				Interactive:    false,
			})
			f.EXPECT().Client().Return(mocks.NewMockClient(t), nil)
			f.EXPECT().GetWorkspaceID().Return(tt.expected.Workspace, nil)

			ted := mocks.NewMockTimeEntryDefaults(t)
			ted.EXPECT().Read().Return(tt.current, nil)
			ted.EXPECT().Write(tt.expected).Return(nil)
			f.EXPECT().TimeEntryDefaults().Return(ted)

			called := false
			var result defaults.DefaultTimeEntry
			cmd := set.NewCmdSet(f, func(_ OutputFlags, _ io.Writer,
				dte defaults.DefaultTimeEntry) error {
				called = true
				result = dte
				return nil
			})

			cmd.SilenceUsage = true
			cmd.SetArgs(tt.args)
			_, err := cmd.ExecuteC()

			assert.NoError(t, err, "should not have failed")
			assert.True(t, called)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewCmdSet_ShouldUpdateDefaultsFile_OnlyByFlags(t *testing.T) {
	tts := []struct {
		name     string
		args     []string
		current  defaults.DefaultTimeEntry
		expected defaults.DefaultTimeEntry
	}{
		{
			name: "no arguments, no changes",
			args: []string{},
			current: defaults.DefaultTimeEntry{
				Workspace: "w1", ProjectID: "p1"},
			expected: defaults.DefaultTimeEntry{
				Workspace: "w1", ProjectID: "p1"},
		},
		{
			name: "all arguments",
			args: []string{
				"-p=p2",
				"--task=t2",
				"-T=tg1", "-T=tg2",
				"--billable",
			},
			expected: defaults.DefaultTimeEntry{
				Workspace: "w2",
				ProjectID: "p2",
				TaskID:    "t2",
				Billable:  &bTrue,
				TagIDs:    []string{"tg1", "tg2"},
			},
		},
		{
			name: "not billable",
			args: []string{"--not-billable"},
			current: defaults.DefaultTimeEntry{
				Workspace: "w2",
				ProjectID: "p2",
				TaskID:    "t2",
				Billable:  &bTrue,
				TagIDs:    []string{"tg1", "tg2"},
			},
			expected: defaults.DefaultTimeEntry{
				Workspace: "w2",
				ProjectID: "p2",
				TaskID:    "t2",
				Billable:  &bFalse,
				TagIDs:    []string{"tg1", "tg2"},
			},
		},
	}

	for i := range tts {
		tt := &tts[i]
		t.Run(tt.name, func(t *testing.T) {
			f := mocks.NewMockFactory(t)
			f.EXPECT().Config().Return(&mocks.SimpleConfig{
				AllowNameForID: false,
				Interactive:    false,
			})
			f.EXPECT().Client().Return(mocks.NewMockClient(t), nil)
			f.EXPECT().GetWorkspaceID().Return(tt.expected.Workspace, nil)

			ted := mocks.NewMockTimeEntryDefaults(t)
			ted.EXPECT().Read().Return(tt.current, nil)
			ted.EXPECT().Write(tt.expected).Return(nil)
			f.EXPECT().TimeEntryDefaults().Return(ted)

			called := false
			var result defaults.DefaultTimeEntry
			cmd := set.NewCmdSet(f, func(_ OutputFlags, _ io.Writer,
				dte defaults.DefaultTimeEntry) error {
				called = true
				result = dte
				return nil
			})

			cmd.SilenceUsage = true
			cmd.SetArgs(tt.args)
			_, err := cmd.ExecuteC()

			assert.NoError(t, err, "should not have failed")
			assert.True(t, called)
			assert.Equal(t, tt.expected, result)
		})
	}
}
