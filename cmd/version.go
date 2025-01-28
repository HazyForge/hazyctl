package cmd

import (
    "fmt"
    "hazyctl/pkg/version"
    "github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the version",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("Version: %s\n", version.Version)
        fmt.Printf("Git Commit:    %s\n", version.Commit)
        fmt.Printf("Build Date:    %s\n", version.Date)
    },
}


func init() {
    rootCmd.AddCommand(versionCmd)
}