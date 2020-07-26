/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
This file contains variables and functions to handle the default
command line behavior (interactive mode) or to pass sub-commands
to one of the cobra commands defined in other files within this
package, such as add_note.go.
*/

package cmd

import (
	"fmt"
	"io"
	"memory/app"
	"memory/app/config"
	"memory/app/persist"
	"os"
	"strings"

	"github.com/chzyer/readline"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var memoryHome string
var memoryHomeName = ".memory"
var settingsFile = "settings.yml"

// completer dictates the readline tab completion options
var completer = readline.NewPrefixCompleter(
	readline.PcItem("add-event"),
	readline.PcItem("add-note"),
	readline.PcItem("add-person"),
	readline.PcItem("add-place"),
	readline.PcItem("add-thing"),
	readline.PcItem("detail"),
	readline.PcItem("ls",
		readline.PcItem("--types"),
		readline.PcItem("--tags"),
		readline.PcItem("--contains"),
		readline.PcItem("--start-with"),
	),
	readline.PcItem("delete"),
	readline.PcItem("rename"),
	readline.PcItem("edit"),
)

// filterInput allows certain keys to be intercepted during readline
func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// the rl library provides bash-like completion in interactive mode
var rl *readline.Instance

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "memory",
	Short: "A database for human experience",
	Long: `Memory is a CLI application that captures and stores the people, places, 
things and events that make up human experience and helps you link them together 
in interesting ways.`,
	Run: func(cmd *cobra.Command, args []string) {
		welcomeMessage()
		// readline setup
		var err error
		rl, err = readline.NewEx(&readline.Config{
			Prompt:              config.Prompt,
			HistoryFile:         config.HistoryPath(),
			AutoComplete:        completer,
			InterruptPrompt:     "^C",
			EOFPrompt:           "exit",
			HistorySearchFold:   true,
			FuncFilterInputRune: filterInput,
		})
		if err != nil {
			panic(err)
		}
		defer rl.Close()
		// input loop
		for {
			line, err := rl.Readline()
			if err == readline.ErrInterrupt {
				if len(line) == 0 {
					break
				} else {
					continue
				}
			} else if err == io.EOF {
				break
			}
			line = strings.TrimSpace(line)
			switch {
			case line == "exit" || line == "quit":
				os.Exit(0)
			case strings.HasPrefix(line, "add-note"):
				addNoteInteractive(line[4:]) // in add_note.go
			case line == "ls" || strings.HasPrefix(line, "ls "):
				args := ""
				if len(line) > 3 {
					args = line[3:]
				}
				lsInteractive(args) // in ls.go
			case strings.HasPrefix(line, "detail ") || line == "detail":
				if line == "detail" {
					fmt.Println("Usage: detail [name]")
					continue
				}
				line = strings.TrimSpace(line[7:])
				detailInteractive(line)
			case strings.HasPrefix(line, "delete ") || line == "delete":
				if line == "delete" {
					fmt.Println("Usage: delete [name]")
					continue
				}
				line = strings.TrimSpace(line[7:])
				deleteEntryInteractive(line)
			case strings.HasPrefix(line, "edit ") || line == "edit":
				if line == "edit" {
					fmt.Println("Usage: edit [name]")
					continue
				}
				line = strings.TrimSpace(line[5:])
				editInteractive(line)
			default:
				//TODO: Implement help command in interactive mode
				fmt.Printf("Sorry, I don't understand '%s'. Try 'help'.\n", line)
			}
		}
	},
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

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		// Otherwise, save the defaults
		saveAs := memoryHome + slash + settingsFile
		viper.SafeWriteConfigAs(saveAs)
		fmt.Println("Created default settings file:", saveAs)
	}

	// Set app config values from viper
	config.MemoryHome = memoryHome

	//TODO: Add config settings to settings file

	// Initialize app state
	if err := app.Init(); err != nil {
		panic("Failed to initialize application state: " + err.Error())
	}
}

// welcomeMessage personalizes the app with a message tailored to the visitors
// current journey.
//TODO: Flesh out the welcome journey
func welcomeMessage() {
	fmt.Printf("Welcome. You have %d entries under management. "+
		"Type 'help' for assistance.\n", app.EntryCount())
}
