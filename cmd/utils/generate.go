package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/storage"
)

var (
	genLength    int
	genNoSymbols bool
	genSaveKey   string
	genNamespace string
	genNoCopy    bool
)

var GenerateCmd = &cobra.Command{
	GroupID: "utils",
	Use:     "generate",
	Short:   "Generate a secure random password",
	RunE:    runGenerate,
}

func init() {
	GenerateCmd.Flags().IntVarP(&genLength, "length", "l", 24, "password length")
	GenerateCmd.Flags().BoolVar(&genNoSymbols, "no-symbols", false, "alphanumeric only")
	GenerateCmd.Flags().StringVar(&genSaveKey, "save", "", "save as KEY in vault")
	GenerateCmd.Flags().StringVar(&genNamespace, "ns", "default", "namespace for --save")
	GenerateCmd.Flags().BoolVar(&genNoCopy, "no-copy", false, "do not copy to clipboard")
}

func runGenerate(cmd *cobra.Command, _ []string) error {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const symbols = "!@#$%^&*()-_=+[]{}|;:,.<>?"
	charset := letters
	if !genNoSymbols {
		charset += symbols
	}

	if genLength < 8 {
		return fmt.Errorf("length must be at least 8")
	}

	buf := make([]byte, genLength)
	for i := range buf {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return fmt.Errorf("random: %w", err)
		}
		buf[i] = charset[n.Int64()]
	}
	password := string(buf)

	fmt.Println(password)

	if genSaveKey != "" {
		if err := common.ValidateKey(genSaveKey); err != nil {
			return err
		}
		ns := common.ActiveNamespace(cmd, genNamespace)
		err := common.MutateVault(func(vd *storage.VaultData) error {
			if vd.Namespaces[ns] == nil {
				vd.Namespaces[ns] = make(map[string]string)
			}
			vd.Namespaces[ns][genSaveKey] = password
			return nil
		})
		if err != nil {
			return err
		}
		lpath, _ := common.LogPath()
		_ = audit.Append(lpath, audit.OpSet, genSaveKey, ns)
		fmt.Printf("Saved as %s in [%s]\n", genSaveKey, ns)
	}

	if !genNoCopy {
		if err := common.CopyToClipboard(password); err == nil {
			common.ScheduleClipboardClear(30)
			fmt.Println("Copied to clipboard. Will clear in 30 seconds.")
		}
	}

	return nil
}
