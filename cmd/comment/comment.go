package comment

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "comment [resource] [id] [text]",
	Short: "Add a comment to a resource",
	Long:  "Post a comment to a specific resource in OpenProject.",
}
