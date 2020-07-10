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
	readline.PcItem("add",
		readline.PcItem("note"),
		readline.PcItem("event"),
		readline.PcItem("person"),
		readline.PcItem("place"),
		readline.PcItem("thing"),
	),
	readline.PcItem("ls",
		readline.PcItem("--types"),
	),
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
things and events that make up human experience and links them together 
in interesting ways.`,
	Run: func(cmd *cobra.Command, args []string) {
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
			case strings.HasPrefix(line, "add "):
				switch strings.TrimSpace(line[4:]) {
				case "note":
					addNoteInteractive(line[4:]) // in add_note.go
				}
			case line == "ls" || strings.HasPrefix(line, "ls "):
				args := ""
				if len(line) > 3 {
					args = line[3:]
				}
				lsInteractive(args) // in ls.go
			default:
				fmt.Println("Sorry, I don't understand. Try 'help'.")
			}
		}
	},
}

// subPrompt asks for additional info within a command.
func subPrompt(prompt string, validate validator) string {
	rl.HistoryDisable()
	rl.SetPrompt(prompt)
	var err error
	var input = ""
	for {
		input, err = rl.ReadlineWithDefault(input)
		if err != nil {
			break
		}
		if msg := validate(input); msg != "" {
			fmt.Println(msg)
		} else {
			break
		}
	}
	rl.HistoryEnable()
	rl.SetPrompt(config.Prompt)
	return strings.TrimSpace(input)
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
