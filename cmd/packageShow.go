package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/nikoksr/proji/messages"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/pkg/errors"

	"github.com/nikoksr/proji/util"

	"github.com/nikoksr/proji/storage/models"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

type packageShowCommand struct {
	cmd *cobra.Command
}

func newPackageShowCommand() *packageShowCommand {
	var showAll bool

	var cmd = &cobra.Command{
		Use:   "show LABEL [LABEL...]",
		Short: "Show details about one or more packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !showAll && len(args) < 1 {
				return fmt.Errorf("missing package label")
			}

			var labels []string
			if !showAll {
				labels = args
			}
			return showPackages(labels...)
		},
	}
	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all packages")
	return &packageShowCommand{cmd: cmd}
}

func showPackage(preloadedPackage *models.Package, label string) error {
	var err error
	if preloadedPackage == nil {
		preloadedPackage, err = activeSession.storageService.LoadPackage(label)
		if err != nil {
			return errors.Wrap(err, "failed to load package")
		}
	}
	output := os.Stdout
	showBasicInfo(preloadedPackage.Name, preloadedPackage.Label, preloadedPackage.Description)
	showTemplates(output, preloadedPackage.Templates)
	showPlugins(output, preloadedPackage.Plugins)
	return nil
}

func showPackages(labels ...string) error {
	packages, err := activeSession.storageService.LoadPackages(labels...)
	if err != nil {
		return errors.Wrap(err, "failed to load package")
	}
	for _, pkg := range packages {
		err = showPackage(pkg, pkg.Label)
		if err != nil {
			messages.Warningf("failed to show package %s, %s", pkg.Label, err.Error())
		}
	}
	return nil
}

func showBasicInfo(name, label, description string) {
	fmt.Printf("\nName:  %s\n", name)
	fmt.Printf("Label: %s\n", label)
	fmt.Printf("Description: %s\n\n", text.WrapSoft(description, activeSession.maxTableColumnWidth))
}

func showTemplates(out io.Writer, templates []*models.Template) {
	templatesTable := util.NewInfoTable(out)
	templatesTable.SetTitle("TEMPLATES")
	templatesTable.AppendHeader(table.Row{"Destination", "Template Path", "Is File", "Description"})

	for _, template := range templates {
		templatesTable.AppendRow(
			table.Row{
				template.Destination,
				template.Path,
				template.IsFile,
				template.Description,
			},
		)
	}
	templatesTable.Render()
}

func showPlugins(out io.Writer, plugins []*models.Plugin) {
	pluginsTable := util.NewInfoTable(out)
	pluginsTable.SetTitle("PLUGINS")
	pluginsTable.AppendHeader(table.Row{"Path", "Execution Number", "Description"})

	for _, plugin := range plugins {
		pluginsTable.AppendRow(
			table.Row{
				plugin.Path,
				plugin.ExecNumber,
				text.WrapSoft(plugin.Description, activeSession.maxTableColumnWidth),
			},
		)
	}
	pluginsTable.Render()
}
