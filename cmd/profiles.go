/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type CfgService struct {
	Name      string `yaml:"name" mapstructure:"name"`
	Ports     []int  `yaml:"ports" mapstructure:"ports"`
	Context   string `yaml:"context,omitempty" mapstructure:"context"`
	Namespace string `yaml:"namespace,omitempty" mapstructure:"namespace"`
}

type CfgProfileItem struct {
	Name      string       `yaml:"name" mapstructure:"name"`
	Context   string       `yaml:"context,omitempty" mapstructure:"context"`
	Namespace string       `yaml:"namespace,omitempty" mapstructure:"namespace"`
	Services  []CfgService `yaml:"services" mapstructure:"services"`
}

type CfgProfile struct {
	Context        string           `yaml:"context,omitempty" mapstructure:"context"`
	Namespace      string           `yaml:"namespace,omitempty" mapstructure:"namespace"`
	DefaultProfile string           `yaml:"default-profile,omitempty" mapstructure:"default-profile"`
	Profiles       []CfgProfileItem `yaml:"profiles" mapstructure:"profiles"`
}

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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			//cmd.Help()
			//os.Exit(1)
			return fmt.Errorf("unknown sub-command %s", args[0])
		}

		return nil
	},
}

var profilesRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a profile",
	Run: func(cmd *cobra.Command, args []string) {
		profile := cmd.Flag("profile").Value.String()
		runProfile(profile)
	},
}

func init() {
	rootCmd.AddCommand(profilesCmd)
	profilesCmd.Flags().BoolP("init", "i", false, "If enabled, will create a default configuration file")
	profilesCmd.Flags().BoolP("overwrite", "o", false, "If enabled, will overwrite any existing configuration file")

	profilesCmd.AddCommand(profilesRunCmd)
	profilesRunCmd.Flags().StringP("profile", "p", "", "It will run A profile, it will default to the global 'default-profile' configuration")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// profilesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// profilesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runProfile(targetProfile string) {
	fmt.Println("Running `profiles run` targetProfile [", targetProfile, "]")

	profile, err := LoadProfile()
	if err != nil {
		log.Fatalf("unable to read config from file: %v", err)
	}

	if targetProfile == "" {
		targetProfile = profile.DefaultProfile
	}

	if targetProfile == "" {
		log.Fatalf("no default profile found")
	}

	foundIdx := slices.IndexFunc(profile.Profiles, func(i CfgProfileItem) bool {
		fmt.Println("Comparing", i.Name, targetProfile)
		return i.Name == targetProfile
	})

	if foundIdx < 0 {
		log.Fatalf("profile %s not found [%d]", targetProfile, foundIdx)
	}

	printProfileItem(profile.Profiles[foundIdx])
}

func runProfiles(cmd *cobra.Command, isInit bool, overwrite bool) {
	file := viper.ConfigFileUsed()
	fmt.Println("Running `profiles` init? [", isInit, "] overwrite [", overwrite, "file [", file, "]")

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

func initConfigProfile(_ *cobra.Command) {
	ctx := viper.GetString("context")
	ns := viper.GetString("namespace")

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

// keep Global namespace and context
func MergeProfile(newProfile *CfgProfile) CfgProfile {
	currentProfile, err := LoadProfile()
	if err != nil {
		log.Fatalf("unable to read config from file: %v", err)
	}

	if currentProfile.Context == "" {
		currentProfile.Context = newProfile.Context
	}

	if currentProfile.Namespace == "" {
		currentProfile.Namespace = newProfile.Namespace
	}

	if currentProfile.DefaultProfile == "" {
		currentProfile.DefaultProfile = newProfile.DefaultProfile
	}

	// FIXME this is going to override existing profiles (names)
	allProfiles := append(currentProfile.Profiles, newProfile.Profiles...)

	return CfgProfile{
		Context:        currentProfile.Context,
		Namespace:      currentProfile.Namespace,
		DefaultProfile: currentProfile.DefaultProfile,
		Profiles:       allProfiles,
	}

}

func WriteProfile(p *CfgProfile) {
	bs, err := yaml.Marshal(p)
	if err != nil {
		log.Fatalf("unable to marshal config to YAML: %v", err)
	}
	contents := string(bs)
	fmt.Println("Writing config to", viper.ConfigFileUsed())
	fmt.Println(contents)

	if err := viper.ReadConfig(strings.NewReader(contents)); err != nil {
		log.Fatalf("unable to read config to file: %v", err)
	}

	if err := viper.WriteConfig(); err != nil {
		log.Fatalf("unable to merge config to file: %v", err)
	}

}

func LoadProfile() (CfgProfile, error) {
	file := viper.ConfigFileUsed()
	fmt.Println("Loading profiles from ", file)

	var profile CfgProfile

	if err := viper.Unmarshal(&profile); err != nil {
		return CfgProfile{}, fmt.Errorf("could not parse file \"%s\". Error: %s", file, err)
	}

	return profile, nil
}

func printProfiles(cfg CfgProfile) {
	fmt.Printf("Context: %s\n", cfg.Context)
	fmt.Printf("Namespace: %s\n", cfg.Namespace)

	for _, p := range cfg.Profiles {
		printProfileItem(p)
	}
}

func printProfileItem(item CfgProfileItem) {
	fmt.Printf("Profile: %s\n", item.Name)
	fmt.Printf("  Context: %s, Namespace: %s\n", item.Context, item.Namespace)
	fmt.Println("  Services:")
	for _, s := range item.Services {
		fmt.Printf("    - Name: %s, Ports: %v", s.Name, s.Ports)
		if s.Context != "" {
			fmt.Printf(", Context: %s", s.Context)
		}
		fmt.Println()
	}
	fmt.Println()

}
