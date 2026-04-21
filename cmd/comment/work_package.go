package comment

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/opf/openproject-cli/components/printer"
	"github.com/opf/openproject-cli/components/resources/work_packages"
)

var commentWorkPackageCmd = &cobra.Command{
	Use:   "workpackage [id] [text]",
	Short: "Comment on a work package",
	Long:  "Add a markdown comment to a work package by its ID.",
	Run:   commentWorkPackage,
}

func commentWorkPackage(_ *cobra.Command, args []string) {
	if len(args) < 2 {
		printer.ErrorText(fmt.Sprintf("Expected 2 arguments [id] [text], but got %d", len(args)))
		return
	}

	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		printer.ErrorText(fmt.Sprintf("'%s' is an invalid work package id. Must be a number.", args[0]))
		return
	}

	text := args[1]
	activity, err := work_packages.Comment(id, text)
	if err != nil {
		printer.Error(err)
		return
	}

	printer.Info("Comment added successfully.")
	printer.Activity(activity)
}

func init() {
	RootCmd.AddCommand(commentWorkPackageCmd)
}
