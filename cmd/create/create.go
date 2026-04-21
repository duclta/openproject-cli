package create

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "create [resource]",
	Short: "Creates a specific resource",
	Long:  "Create a specific resource in OpenProject",
}

func init() {
	createWorkPackageCmd.Flags().Uint64VarP(
		&projectId,
		"project",
		"p",
		0,
		"Project ID to create the work package in",
	)
	_ = createWorkPackageCmd.MarkFlagRequired("project")

	createWorkPackageCmd.Flags().BoolVarP(
		&shouldOpenWorkPackageInBrowser,
		"open",
		"o",
		false,
		"Open the created work package in the default browser",
	)

	createWorkPackageCmd.Flags().StringVarP(
		&typeFlag,
		"type",
		"t",
		"",
		"Change the work package type",
	)

	createWorkPackageCmd.Flags().StringVar(
		&descriptionFlag,
		"description",
		"",
		"Work package description (markdown supported)",
	)

	createWorkPackageCmd.Flags().StringVar(
		&parentFlag,
		"parent",
		"",
		"Parent work package ID",
	)

	createWorkPackageCmd.Flags().StringVar(
		&versionFlag,
		"version",
		"",
		"Version name or ID",
	)

	createWorkPackageCmd.Flags().StringVar(
		&startDateFlag,
		"start-date",
		"",
		"Start date (YYYY-MM-DD)",
	)

	createWorkPackageCmd.Flags().StringVar(
		&dueDateFlag,
		"due-date",
		"",
		"Due date (YYYY-MM-DD)",
	)

	RootCmd.AddCommand(createWorkPackageCmd)
}
