package cmd

import (
	"os"

	"github.com/nikoksr/proji/util"
	"github.com/pkg/errors"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

type packageListCommand struct {
	cmd *cobra.Command
}

func newPackageListCommand() *packageListCommand {
	var cmd = &cobra.Command{
		Use:                   "ls",
		Short:                 "List packages",
		DisableFlagsInUseLine: true,
		Args:                  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listPackages()
		},
	}
	return &packageListCommand{cmd: cmd}
}

func listPackages() error {
	packages, err := activeSession.storageService.LoadPackages()
	if err != nil {
		return errors.Wrap(err, "failed to load all packages")
	}

	packagesTable := util.NewInfoTable(os.Stdout)
	packagesTable.AppendHeader(table.Row{"Name", "Label"})

	for _, pkg := range packages {
		if pkg.IsDefault {
			continue
		}
		packagesTable.AppendRow(table.Row{pkg.Name, pkg.Label})
	}
	packagesTable.Render()
	return nil
}
