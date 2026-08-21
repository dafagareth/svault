package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"svault/cmd/auth"
	"svault/cmd/common"
	"svault/cmd/env"
	"svault/cmd/ns"
	"svault/cmd/secrets"
	"svault/cmd/system"
	"svault/cmd/utils"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:          "svault",
	Short:        "Local encrypted secret vault",
	SilenceUsage: true,
}

func init() {
	rootCmd.AddGroup(
		&cobra.Group{ID: "auth", Title: "Session & Authentication:"},
		&cobra.Group{ID: "secrets", Title: "Secret Operations:"},
		&cobra.Group{ID: "env", Title: "Environment & Integration:"},
		&cobra.Group{ID: "ns", Title: "Namespaces:"},
		&cobra.Group{ID: "utils", Title: "Utilities & Password Generation:"},
		&cobra.Group{ID: "system", Title: "Maintenance & Diagnostics:"},
	)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(
		auth.InitCmd,
		auth.UnlockCmd,
		auth.LockCmd,
		auth.StatusCmd,
		secrets.SetCmd,
		secrets.GetCmd,
		secrets.DeleteCmd,
		secrets.EditCmd,
		secrets.ListCmd,
		secrets.SearchCmd,
		secrets.RenameCmd,
		secrets.MoveCmd,
		env.ExecCmd,
		env.EnvCmd,
		env.ExportCmd,
		env.ImportCmd,
		env.CheckCmd,
		ns.NsCmd,
		ns.UseCmd,
		ns.DiffCmd,
		utils.GenerateCmd,
		utils.CopyCmd,
		utils.OpenCmd,
		system.RotateCmd,
		system.BackupCmd,
		system.RestoreCmd,
		system.LogCmd,
		system.DoctorCmd,
		system.InfoCmd,
		system.CompletionCmd,
		system.VersionCmd,
	)
}

func Execute() {
	system.Version = version
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("svault v{{.Version}}\n")
	exempt := map[string]bool{
		"init":       true,
		"help":       true,
		"completion": true,
		"version":    true,
		"doctor":     true,
	}
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if exempt[cmd.Name()] {
			return nil
		}
		vpath, err := common.VaultPath()
		if err != nil {
			return err
		}
		if _, err := os.Stat(vpath); os.IsNotExist(err) {
			return fmt.Errorf("no vault found.\n\nGet started:\n  svault init        create a new encrypted vault\n  svault --help      show all commands\n\nVault will be stored at ~/.svault/vault.enc")
		}
		return nil
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
