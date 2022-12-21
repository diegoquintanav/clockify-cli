package set

import (
	"io"

	"github.com/lucassabreu/clockify-cli/pkg/cmd/time-entry/util/defaults"
	"github.com/lucassabreu/clockify-cli/pkg/cmdutil"
	. "github.com/lucassabreu/clockify-cli/pkg/output/defaults"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewCmdSet sets the default parameters for time entries in the current folder
func NewCmdSet(
	f cmdutil.Factory,
	get func() (defaults.DefaultTimeEntry, error),
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
			d, _ := get()

			return report(of, cmd.OutOrStdout(), d)
		},
	}
	return cmd
}
