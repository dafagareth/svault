package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionShort bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the svault version",
	Run: func(_ *cobra.Command, _ []string) {
		if versionShort {
			fmt.Println(version)
		} else {
			fmt.Printf("svault v%s\n", version)
		}
	},
}

func init() {
	versionCmd.Flags().BoolVar(&versionShort, "short", false, "print version number only")
	rootCmd.AddCommand(versionCmd)
}
