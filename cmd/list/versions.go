package list

import (
	"github.com/spf13/cobra"

	"github.com/opf/openproject-cli/components/printer"
	"github.com/opf/openproject-cli/components/resources/projects"
)

var versionsProjectId uint64

var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Lists project versions",
	Long:  "Get a list of all versions of the specified project.",
	Run:   listVersions,
}

func initVersionsFlags() {
	versionsCmd.Flags().Uint64VarP(
		&versionsProjectId,
		"project-id",
		"p",
		0,
		"Show versions of the specified projectId",
	)
	_ = versionsCmd.MarkFlagRequired("project-id")
}

func listVersions(_ *cobra.Command, _ []string) {
	if all, err := projects.AvailableVersions(versionsProjectId); err == nil {
		printer.Versions(all)
	} else {
		printer.Error(err)
	}
}
