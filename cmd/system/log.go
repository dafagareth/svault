package system

import (
	"fmt"
	"os"
	"strings"
	"svault/cmd/common"

	"github.com/spf13/cobra"
)

var logTail int

var LogCmd = &cobra.Command{
	GroupID: "system",
	Use:     "log",
	Short:   "Show the vault audit log",
	RunE:    runLog,
}

func init() {
	LogCmd.Flags().IntVarP(&logTail, "tail", "n", 0, "show last N entries (0 = all)")
}

func runLog(_ *cobra.Command, _ []string) error {
	lpath, err := common.LogPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(lpath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("(empty)")
			return nil
		}
		return fmt.Errorf("read log: %w", err)
	}
	if len(data) == 0 {
		fmt.Println("(empty)")
		return nil
	}
	if logTail > 0 {
		lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		if logTail < len(lines) {
			lines = lines[len(lines)-logTail:]
		}
		fmt.Println(strings.Join(lines, "\n"))
		return nil
	}
	fmt.Print(string(data))
	return nil
}
