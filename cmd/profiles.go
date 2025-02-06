/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// profilesCmd represents the profiles command
var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		isInit := cmd.Flag("init").Value.String() == "true"
		overwrite := cmd.Flag("overwrite").Value.String() == "true"

		runProfiles(cmd, isInit, overwrite)
	},
}

func init() {
	rootCmd.AddCommand(profilesCmd)
	profilesCmd.Flags().BoolP("init", "i", false, "If enabled, will create a default configuration file")
	profilesCmd.Flags().BoolP("overwrite", "o", false, "If enabled, will overwrite any existing configuration file")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// profilesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// profilesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type CfgService struct {
	Name    string `yaml:"name" mapstructure:"name"`
	Ports   []int  `yaml:"ports" mapstructure:"ports"`
	Context string `yaml:"context,omitempty" mapstructure:"context"`
}

type CfgProfileItem struct {
	Name      string       `yaml:"name" mapstructure:"name"`
	Context   string       `yaml:"context,omitempty" mapstructure:"context"`
	Namespace string       `yaml:"namespace,omitempty" mapstructure:"namespace"`
	Services  []CfgService `yaml:"services" mapstructure:"services"`
}

type CfgProfile struct {
	Context   string           `yaml:"context,omitempty" mapstructure:"context"`
	Namespace string           `yaml:"namespace,omitempty" mapstructure:"namespace"`
	Profiles  []CfgProfileItem `yaml:"profiles" mapstructure:"profiles"`
}

func runProfiles(cmd *cobra.Command, isInit bool, overwrite bool) {
	file := viper.ConfigFileUsed()
	fmt.Println("profiles called isInit", isInit, file, overwrite)

	if isInit && overwrite {
		initConfigProfile(cmd)
		return
	}

	if file != "" && isInit && !overwrite {
		fmt.Printf("PROFILE FILE FOUND at %s, Try using --init with --overwrite flag\n", file)
		return
	}

	if file == "" {
		if !isInit {
			fmt.Println("NO PROFILES FOUND, Try using --init flag")
		} else {
			initConfigProfile(cmd)
		}
	} else {
		fmt.Println("Config file:", file)
		p, _ := LoadProfile()
		printProfiles(p)
	}
}

func initConfigProfile(cmd *cobra.Command) {
	ns := cmd.Flag("namespace").Value.String()
	ctx := cmd.Flag("context").Value.String()

	fmt.Println("Creating default configuration file ns ", ns, " ctx ", ctx)
	WriteProfile(&CfgProfile{
		Namespace: ns,
		Context:   ctx,
		Profiles: []CfgProfileItem{
			{
				Name:      "foo",
				Namespace: "bar",
				Services: []CfgService{
					{
						Name:  "typhoon",
						Ports: []int{9000},
					},
					{
						Name:  "couchdb",
						Ports: []int{5984},
					},
				},
			},
		},
	})
}

func WriteProfile(p *CfgProfile) {
	bs, err := yaml.Marshal(p)
	if err != nil {
		log.Fatalf("unable to marshal config to YAML: %v", err)
	}

	contents := string(bs)
	fmt.Println("Writing config to", viper.ConfigFileUsed())
	fmt.Println(contents)

	viper.ReadConfig(strings.NewReader(contents))
	viper.WriteConfig()
}

func LoadProfile() (CfgProfile, error) {
	file := viper.ConfigFileUsed()
	fmt.Println("Loading profiles from ", file)

	var profile CfgProfile

	if err := viper.Unmarshal(&profile); err != nil {
		return CfgProfile{}, fmt.Errorf("could not parse file \"%s\". Error: %s", file, err)
	}

	return profile, nil

	// fmt.Println("Loading profiles")
	// var profiles []Profile
	// // Unmarshal the "profiles" key from the YAML into a slice of Profile structs.
	// if err := viper.UnmarshalKey("profiles", &profiles); err != nil {
	// 	fmt.Printf("Error parsing profiles: %v\n", err)
	// 	return CfgProfile{}, nil
	// }

	// return CfgProfile{
	// 	Namespace: viper.GetString("namespace"),
	// 	Context:   viper.GetString("context"),
	// 	Profiles:  profiles,
	// }, nil
}

func printProfiles(cfg CfgProfile) {
	fmt.Printf("Context: %s\n", cfg.Context)
	fmt.Printf("Namespace: %s\n", cfg.Namespace)

	// Iterate over each profile and print its details.
	for _, p := range cfg.Profiles {
		fmt.Printf("Profile: %s\n", p.Name)
		fmt.Printf("  Context: %s, Namespace: %s\n", p.Context, p.Namespace)
		fmt.Println("  Services:")
		for _, s := range p.Services {
			fmt.Printf("    - Name: %s, Ports: %v", s.Name, s.Ports)
			if s.Context != "" {
				fmt.Printf(", Context: %s", s.Context)
			}
			fmt.Println()
		}
		fmt.Println()
	}
}
