package system

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "dev"
var versionShort bool

var VersionCmd = &cobra.Command{
	GroupID: "system",
	Use:     "version",
	Short:   "Print the svault version",
	Run: func(_ *cobra.Command, _ []string) {
		if versionShort {
			fmt.Println(Version)
		} else {
			fmt.Printf("svault v%s\n", Version)
		}
	},
}

func init() {
	VersionCmd.Flags().BoolVar(&versionShort, "short", false, "print version number only")
}
