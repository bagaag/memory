/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/
package cmd

import (
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"memory/app"
	"memory/app/config"
	"memory/app/persist"
	"os"
)

var memoryHome string
var memoryHomeName = ".memory"
var settingsFile = "settings.yml"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "memory",
	Short: "A database for human experience",
	Long: `Memory is a CLI application that captures and stores the people, places, 
things and events that make up human experience and links them together 
in interesting ways.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//TODO: This isn't getting set when specified
	rootCmd.PersistentFlags().StringVar(&memoryHome, "home", "", "Config and save folder, default is $HOME/.memory")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// Set default settings in viper
	viper.SetDefault("bagaag", "kneeg")

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println("Could not find home directory:", err)
		// Fail gracefully and use current working directory if home can't be located
		if home, err = os.Getwd(); err != nil {
			fmt.Println("Could not find working directory:", err)
			home = "."
		}
	}

	slash := string(os.PathSeparator)

	// Set home location if not set in flag
	if memoryHome == "" {
		memoryHome = home + slash + memoryHomeName
	}

	// Create MemoryHome folder if it doesn't exist
	if !persist.PathExists(memoryHome) {
		os.Mkdir(memoryHome, os.ModeDir+0700)
		fmt.Println("Created save folder:", memoryHome)
	}

	// Populate viper settings
	viper.SetConfigFile(memoryHome + slash + settingsFile)
	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err == nil {
		// If a config file is found, read it in.
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println(err)
		// Otherwise, save the defaults
		saveAs := memoryHome + slash + settingsFile
		viper.SafeWriteConfigAs(saveAs)
		fmt.Println("Created default settings file:", saveAs)
	}

	// Set app config values from viper
	config.MemoryHome = memoryHome

	// Initialize app state
	if err := app.Init(); err != nil {
		panic("Failed to initialize application state: " + err.Error())
	}
}

// Shared function for use by commands to save data to config.Savepath.
func save() {
	if err := app.Save(); err != nil {
		fmt.Println("Failed to save data:", err)
	}
}
