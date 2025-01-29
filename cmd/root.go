/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hazyforge/hazyctl/cmd/secret"
	"gopkg.in/yaml.v3"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hazyctl",
	Short: "A brief description of your application",
	Long:  `this application is a helper for many things`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(secret.SecretCmd)
}
func getConfigDir() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return filepath.Join(home, ".local", "share", "hazyctl")
}

var cfgFile = ""
var ConfigDir = getConfigDir()

func initConfig() {

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(ConfigDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AutomaticEnv()
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		fmt.Println("Creating default config file...")
		defaultConfigStruct := Config{
			Azure: AzureConfig{
				Subscription: "",
				Migrate: MigrateConfig{
					Source:      "",
					Destination: "",
				},
				Export: ExportConfig{
					Name:   "",
					Output: "secrets.json",
				},
			},
		}

		defaultConfig, err := yaml.Marshal(defaultConfigStruct)
		if err != nil {
			fmt.Println("Error marshalling default config:", err)
			os.Exit(1)
		}

		err = os.MkdirAll(ConfigDir, 0755)
		if err != nil {
			fmt.Println("Error creating config directory:", err)
			os.Exit(1)
		}

		err = os.WriteFile(filepath.Join(ConfigDir, "config.yaml"), defaultConfig, 0644)
		if err != nil {
			fmt.Println("Error writing default config file:", err)
			os.Exit(1)
		}

		fmt.Println("Default config file created at", filepath.Join(ConfigDir, "config.yaml"))
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("Error reading config file:", err)
			os.Exit(1)
		}
	}
}
