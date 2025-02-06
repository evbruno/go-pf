package cmd

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/evbruno/go-pf/lib"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	quiet       bool
	save        bool
	profileName string
)

// discoverCmd represents the discover command
var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover all services (and its ports) in a give kubernetes + namespace",
	Long: `Discover all services (and its ports) in a give kubernetes + namespace.

It will report if any ports are duplicated.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := viper.GetString("context")
		ns := viper.GetString("namespace")

		quiet = cmd.Flag("quiet").Value.String() == "true"
		save = cmd.Flag("save").Value.String() == "true"
		profileName = cmd.Flag("name").Value.String()

		fmt.Println("Discovering services")
		fmt.Println("Context      :", ctx)
		fmt.Println("Namespace    :", ns)
		fmt.Println("Quiet        :", quiet)
		fmt.Println("Save         :", save)
		fmt.Println("ProfileName  :", profileName)
		fmt.Println()

		runDiscover(ctx, ns)
	},
}

func init() {
	rootCmd.AddCommand(discoverCmd)
	discoverCmd.Flags().BoolP("quiet", "q", false, "When quietly, only errors are reported but exit status is always 0 (also allows to save this profile)")
	discoverCmd.Flags().BoolP("save", "s", false, "Save the current discovered service to profile")
	discoverCmd.Flags().StringP("name", "", "", "If present, will save the profile with the given name")
}

func runDiscover(ctx string, ns string) {
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

	if save {
		if len(svcs) == 0 {
			fmt.Println("No services found. Skipping profile creation.")
			return
		}

		cfgSvcs := []CfgService{}
		for _, svc := range svcs {
			ports := []int{}
			for _, port := range svc.Ports {
				portI, err := strconv.Atoi(port)
				if err != nil {
					fmt.Println("Skipping port ", port, err)
				} else {
					ports = append(ports, portI)
				}
			}

			cfgSvcs = append(cfgSvcs, CfgService{
				Name:      svc.Name,
				Context:   svc.Context,
				Namespace: svc.Namespace,
				Ports:     ports,
			})
		}

		newName := profileName
		if newName == "" {
			r := uint64(time.Now().UnixNano())
			newName = fmt.Sprintf("new-profile-%d", r)
		}

		fmt.Println("---")
		fmt.Println("Creating profile")
		fmt.Println(newName)
		fmt.Println(len(cfgSvcs), "new services")
		fmt.Println("---")

		newProfile := &CfgProfile{
			DefaultProfile: newName,
			Profiles: []CfgProfileItem{
				{
					Name:      newName,
					Namespace: ns,
					Context:   ctx,
					Services:  cfgSvcs,
				},
			},
		}

		updatedProfile := MergeProfile(newProfile)

		WriteProfile(&updatedProfile)
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
