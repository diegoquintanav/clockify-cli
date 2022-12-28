package set

import (
	"io"

	"github.com/lucassabreu/clockify-cli/pkg/cmd/time-entry/util"
	"github.com/lucassabreu/clockify-cli/pkg/cmd/time-entry/util/defaults"
	"github.com/lucassabreu/clockify-cli/pkg/cmdcompl"
	"github.com/lucassabreu/clockify-cli/pkg/cmdcomplutil"
	"github.com/lucassabreu/clockify-cli/pkg/cmdutil"
	. "github.com/lucassabreu/clockify-cli/pkg/output/defaults"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewCmdSet sets the default parameters for time entries in the current folder
func NewCmdSet(
	f cmdutil.Factory,
	report func(OutputFlags, io.Writer, defaults.DefaultTimeEntry) error,
) *cobra.Command {
	if report == nil {
		panic(errors.New("report parameter should not be nil"))
	}

	short := "Sets the default parameters for the current folder"
	of := OutputFlags{}
	cmd := &cobra.Command{
		Use:   "set",
		Short: short,
		Long: short + "\n" +
			"The parameters will be saved in the current working directory " +
			"in the file " + defaults.DEFAULT_FILENAME + ".yaml",
		Example: "",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.XorFlagSet(
				cmd.Flags(), "billable", "not-billable"); err != nil {
				return err
			}

			d, _ := f.TimeEntryDefaults().Read()

			c, err := f.Client()
			if err != nil {
				return err
			}

			d.Workspace, _ = f.GetWorkspaceID()
			t, err := util.Do(
				util.TimeEntryDTO{
					Workspace: d.Workspace,
					ProjectID: d.ProjectID,
					TaskID:    d.TaskID,
					TagIDs:    d.TagIDs,
					Billable:  d.Billable,
				},
				util.FillTimeEntryWithFlags(cmd.Flags()),
				util.GetAllowNameForIDsFn(f.Config(), c),
			)
			if err != nil {
				return err
			}

			d = defaults.DefaultTimeEntry{
				Workspace: t.Workspace,
				ProjectID: t.ProjectID,
				TaskID:    t.TaskID,
				Billable:  t.Billable,
				TagIDs:    t.TagIDs,
			}
			if err = f.TimeEntryDefaults().Write(d); err != nil {
				return err
			}

			return report(of, cmd.OutOrStdout(), d)
		},
	}

	cmd.Flags().BoolP("billable", "b", false,
		"time entry should be billable by default")
	cmd.Flags().BoolP("not-billable", "n", false,
		"time entry should not be billable by default")
	cmd.Flags().String("task", "", "default task")
	_ = cmdcompl.AddSuggestionsToFlag(cmd, "task",
		cmdcomplutil.NewTaskAutoComplete(f, true))

	cmd.Flags().StringSliceP("tag", "T", []string{},
		"add tags be used by default")
	_ = cmdcompl.AddSuggestionsToFlag(cmd, "tag",
		cmdcomplutil.NewTagAutoComplete(f))

	cmd.Flags().StringP("project", "p", "", "project to used by default")
	_ = cmdcompl.AddSuggestionsToFlag(cmd, "project",
		cmdcomplutil.NewProjectAutoComplete(f))

	return cmd
}
