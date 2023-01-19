package set

import (
	"io"

	"github.com/lucassabreu/clockify-cli/api"
	"github.com/lucassabreu/clockify-cli/pkg/cmd/time-entry/util/defaults"
	"github.com/lucassabreu/clockify-cli/pkg/cmdcompl"
	"github.com/lucassabreu/clockify-cli/pkg/cmdcomplutil"
	"github.com/lucassabreu/clockify-cli/pkg/cmdutil"
	. "github.com/lucassabreu/clockify-cli/pkg/output/defaults"
	"github.com/lucassabreu/clockify-cli/pkg/search"
	"github.com/lucassabreu/clockify-cli/strhlp"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

			d, err := f.TimeEntryDefaults().Read()
			if err != nil && err != defaults.DefaultsFileNotFoundErr {
				return err
			}

			n, changed := readFlags(d, cmd.Flags())

			if n.Workspace, err = f.GetWorkspaceID(); err != nil {
				return err
			}

			if changed || d.Workspace != n.Workspace {
				if n.TaskID != "" && n.ProjectID == "" {
					return errors.New("can't set task without project")
				}

				c, err := f.Client()
				if err != nil {
					return err
				}

				if f.Config().IsAllowNameForID() {
					if n, err = updateIDsByNames(
						c, n, f.Config()); err != nil {
						return err
					}
				}

				if f.Config().IsInteractive() {
					if n, err = ask(n, f.Config(), c); err != nil {
						return err
					}
				}

				if !f.Config().IsAllowNameForID() {
					if err = checkIDs(c, n); err != nil {
						return err
					}
				}
			}

			if err = f.TimeEntryDefaults().Write(n); err != nil {
				return err
			}

			return report(of, cmd.OutOrStdout(), n)
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

func readFlags(
	d defaults.DefaultTimeEntry,
	f *pflag.FlagSet,
) (defaults.DefaultTimeEntry, bool) {
	changed := false
	if f.Changed("project") {
		d.ProjectID, _ = f.GetString("project")
		changed = true
	}

	if f.Changed("task") {
		d.TaskID, _ = f.GetString("task")
		changed = true
	}

	if f.Changed("tag") {
		d.TagIDs, _ = f.GetStringSlice("tag")
		d.TagIDs = strhlp.Unique(d.TagIDs)
		changed = true
	}

	if f.Changed("billable") {
		b := true
		d.Billable = &b
		changed = true
	} else if f.Changed("not-billable") {
		b := false
		d.Billable = &b
		changed = true
	}

	return d, changed
}

func checkIDs(c api.Client, d defaults.DefaultTimeEntry) error {
	if d.ProjectID != "" {
		p, err := c.GetProject(api.GetProjectParam{
			Workspace: d.Workspace,
			ProjectID: d.ProjectID,
			Hydrate:   d.TaskID != "",
		})

		if err != nil {
			return err
		}

		if d.TaskID != "" {
			found := false
			for i := range p.Tasks {
				if p.Tasks[i].ID == d.TaskID {
					found = true
					break
				}
			}

			if !found {
				return errors.New(
					"can't find task with ID \"" + d.TaskID +
						"\" on project \"" + d.ProjectID + "\"")
			}
		}
	} else if d.TaskID != "" {
		return errors.New("task can't be set without a project")
	}

	archived := false
	tags, err := c.GetTags(api.GetTagsParam{
		Workspace:       d.Workspace,
		Archived:        &archived,
		PaginationParam: api.AllPages(),
	})
	if err != nil {
		return err
	}

	ids := make([]string, len(tags))
	for i := range tags {
		ids[i] = tags[i].ID
	}

	for _, id := range d.TagIDs {
		if !strhlp.InSlice(id, ids) {
			return errors.Errorf("can't find tag with ID \"%s\"", id)
		}
	}

	return nil
}

func updateIDsByNames(
	c api.Client, d defaults.DefaultTimeEntry, cnf cmdutil.Config) (
	defaults.DefaultTimeEntry,
	error,
) {
	var err error
	if d.ProjectID != "" {
		d.ProjectID, err = search.GetProjectByName(c, d.Workspace, d.ProjectID)
		if err != nil {
			d.ProjectID = ""
			d.TaskID = ""
			if !cnf.IsInteractive() {
				return d, err
			}
		}
	}

	if d.TaskID != "" {
		d.TaskID, err = search.GetTaskByName(c, api.GetTasksParam{
			Workspace: d.Workspace,
			ProjectID: d.ProjectID,
			Active:    true,
		}, d.TaskID)
		if err != nil && !cnf.IsInteractive() {
			return d, err
		}
	}

	if len(d.TagIDs) > 0 {
		d.TagIDs, err = search.GetTagsByName(
			c, d.Workspace, !cnf.IsAllowArchivedTags(), d.TagIDs)
		if err != nil && !cnf.IsInteractive() {
			return d, err
		}
	}

	return d, nil
}

func ask(d defaults.DefaultTimeEntry, cnf cmdutil.Config, c api.Client) (
	defaults.DefaultTimeEntry,
	error,
) {

	return d, nil
}
