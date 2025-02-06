package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/evbruno/go-pf/lib"

	"github.com/spf13/cobra"
)

// discoverCmd represents the discover command
var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover all services (and its ports) in a give kubernetes + namespace",
	Long: `Discover all services (and its ports) in a give kubernetes + namespace.

It will report if any ports are duplicated.`,
	Run: func(cmd *cobra.Command, args []string) {
		ns := cmd.Flag("namespace").Value.String()
		ctx := cmd.Flag("context").Value.String()
		q := cmd.Flag("quiet").Value.String() == "true"
		cfg := cmd.Flag("config").Value.String()

		fmt.Println("Context:", ctx)
		fmt.Println("Namespace:", ns)
		fmt.Println("Quiet:", q)
		fmt.Println("Config:", cfg)

		runDiscover(ctx, ns, q)
	},
}

func init() {
	rootCmd.AddCommand(discoverCmd)
	discoverCmd.Flags().BoolP("quiet", "q", false, "When quietly, only errors are reported but exit status is always 0")
}

func runDiscover(ctx string, ns string, quiet bool) {
	svcs, err := lib.GetKubectlServices(ctx, ns)
	if err != nil {
		panic(err)
	}

	duplicatedPorts := findDuplicatedPorts(svcs)

	fmt.Println("---")

	for _, svc := range svcs {
		fmt.Println(svc.Name, "-->", strings.Join(svc.Ports, " "))
	}

	fmt.Println("---")
	fmt.Printf("%sConflicting ports: %s\n", lib.OkIcon(len(duplicatedPorts) == 0), duplicatedPorts)

	if len(duplicatedPorts) > 0 {
		if !quiet {
			os.Exit(1)
		}
	}

}

func findDuplicatedPorts(svcs []lib.K8SService) []string {
	portsInUse := make(map[string]bool)
	duplicatedPorts := make(map[string]bool)

	for _, svc := range svcs {
		for _, port := range svc.Ports {
			if portsInUse[port] {
				duplicatedPorts[port] = true
			} else {
				portsInUse[port] = true
			}
		}
	}

	var result []string
	for port := range duplicatedPorts {
		result = append(result, port)
	}

	slices.Sort(result)
	return result
}
