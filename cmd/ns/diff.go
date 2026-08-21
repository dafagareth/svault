package ns

import (
	"fmt"
	"sort"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var DiffCmd = &cobra.Command{
	GroupID: "ns",
	Use:     "diff NS1 NS2",
	Short:   "Compare the contents of two namespaces",
	Args:    cobra.ExactArgs(2),
	RunE:    runDiff,
}

func runDiff(_ *cobra.Command, args []string) error {
	ns1Name, ns2Name := args[0], args[1]

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

	ns1 := vd.Namespaces[ns1Name]
	ns2 := vd.Namespaces[ns2Name]

	allKeys := make(map[string]struct{})
	for k := range ns1 {
		allKeys[k] = struct{}{}
	}
	for k := range ns2 {
		allKeys[k] = struct{}{}
	}

	if len(allKeys) == 0 {
		fmt.Printf("Both namespaces [%s] and [%s] are empty\n", ns1Name, ns2Name)
		return nil
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	same, diff := 0, 0
	for _, k := range keys {
		v1, in1 := ns1[k]
		v2, in2 := ns2[k]
		switch {
		case in1 && !in2:
			fmt.Printf("< %-30s  only in [%s]\n", k, ns1Name)
			diff++
		case !in1 && in2:
			fmt.Printf("> %-30s  only in [%s]\n", k, ns2Name)
			diff++
		case v1 != v2:
			fmt.Printf("~ %-30s  value differs\n", k)
			diff++
		default:
			fmt.Printf("= %-30s  same\n", k)
			same++
		}
	}

	fmt.Printf("\n[%s] vs [%s]: %d same, %d differ\n", ns1Name, ns2Name, same, diff)
	return nil
}
