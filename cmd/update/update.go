package update

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "update [resource] [id]",
	Short: "Updates the specific resource",
	Long: `Sends an update to the given resource,
which is identified by its id. The data
to update is determined by the provided
flags.`,
}

func init() {
	addWorkPackageFlags()

	RootCmd.AddCommand(workPackageCmd)
}

func addWorkPackageFlags() {
	workPackageCmd.Flags().StringVarP(
		&actionFlag,
		"action",
		"a",
		"",
		"Executes a custom action on a work package",
	)
	workPackageCmd.Flags().Uint64Var(
		&assigneeFlag,
		"assignee",
		0,
		"Assign a user to the work package",
	)
	workPackageCmd.Flags().StringVar(
		&attachFlag,
		"attach",
		"",
		"Attach a file to the work package",
	)
	workPackageCmd.Flags().StringVar(
		&subjectFlag,
		"subject",
		"",
		"Change the subject of the work package",
	)
	workPackageCmd.Flags().StringVarP(
		&typeFlag,
		"type",
		"t",
		"",
		"Change the work package type",
	)
	workPackageCmd.Flags().StringVar(
		&descriptionFlag,
		"description",
		"",
		"Change the description of the work package (markdown supported)",
	)
	workPackageCmd.Flags().StringVar(
		&statusFlag,
		"status",
		"",
		"Change the status of the work package",
	)
	workPackageCmd.Flags().StringVar(
		&priorityFlag,
		"priority",
		"",
		"Change the priority of the work package",
	)
	workPackageCmd.Flags().StringVar(
		&parentFlag,
		"parent",
		"",
		"Change the parent work package",
	)
	workPackageCmd.Flags().StringVar(
		&versionFlag,
		"version",
		"",
		"Change the version of the work package",
	)
	workPackageCmd.Flags().StringVar(
		&startDateFlag,
		"start-date",
		"",
		"Change the start date (YYYY-MM-DD). Use 'none' to clear.",
	)
	workPackageCmd.Flags().StringVar(
		&dueDateFlag,
		"due-date",
		"",
		"Change the due date (YYYY-MM-DD). Use 'none' to clear.",
	)
}
