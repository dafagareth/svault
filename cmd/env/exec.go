package env

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var execNamespace string

var ExecCmd = &cobra.Command{
	GroupID: "env",
	Use:     "exec COMMAND [ARGS...]",
	Short:   "Run a command with secrets injected as environment variables",
	Args:    cobra.MinimumNArgs(1),
	RunE:    runExecCmd,
}

func init() {
	ExecCmd.Flags().StringVar(&execNamespace, "ns", "default", "source namespace")
}

func runExecCmd(cmd *cobra.Command, args []string) error {
	ns := common.ActiveNamespace(cmd, execNamespace)

	vpath, err := common.VaultPath()
	if err != nil {
		return err
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		return err
	}
	vd, err := storage.ReadVault(vpath, encKey)
	if err != nil {
		return err
	}

	extra := make([]string, 0, len(vd.Namespaces[ns]))
	for k, v := range vd.Namespaces[ns] {
		extra = append(extra, k+"="+v)
	}

	c := exec.Command(args[0], args[1:]...)
	c.Env = append(os.Environ(), extra...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}
