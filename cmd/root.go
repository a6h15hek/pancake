/*
Copyright © 2024 Abhishek M. Yadav <abhishekyadav@duck.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/a6h15hek/pancake/utils"
	"github.com/spf13/cobra"
)

var config utils.Config
var initForce bool

var rootCmd = &cobra.Command{
	Use:   "pancake",
	Short: utils.Description,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(utils.LongDescription)
	},
}

func version() {
	fmt.Println("Pancake " + utils.Version)
}

func editConfig() error {
	configPath, err := utils.ConfigPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("pancake.yml does not exist at %s\nRun 'pancake init' first.\n%s", configPath, utils.ConfigHintEditConfig)
	} else if err != nil {
		return fmt.Errorf("could not check pancake.yml at %s: %w", configPath, err)
	}
	fmt.Printf("Opening pancake.yml at: %s\n", configPath)
	return utils.OpenTextFileInDefaultEditor(configPath)
}

func initCommand(force bool) error {
	fmt.Println("Setup of pancake started...")

	configPath, err := utils.ConfigPath()
	if err != nil {
		return err
	}

	configExists := false
	if stat, statErr := os.Stat(configPath); statErr == nil && !stat.IsDir() {
		configExists = true
	} else if statErr != nil && !os.IsNotExist(statErr) {
		return fmt.Errorf("could not check pancake.yml at %s: %w", configPath, statErr)
	}

	if configExists && !force {
		fmt.Println("pancake.yml already exists. Validating...")
		cfg, err := utils.GetConfig()
		if err != nil {
			return err
		}
		fmt.Printf("  home:     %s\n", cfg.Home)
		fmt.Printf("  projects: %d\n", len(cfg.Projects))
		fmt.Printf("  tools:    %d\n", len(cfg.Tools))
	} else {
		if configExists && force {
			backup := configPath + ".bak"
			if err := os.Rename(configPath, backup); err != nil {
				return fmt.Errorf("could not back up existing pancake.yml to %s: %w", backup, err)
			}
			fmt.Printf("Backed up existing pancake.yml to %s\n", backup)
		}
		if err := os.WriteFile(configPath, []byte(utils.DefaultYMLContent), 0644); err != nil {
			return fmt.Errorf("could not create pancake.yml at %s: %w", configPath, err)
		}
		fmt.Printf("Created pancake.yml at %s\n", configPath)
	}

	cfg, err := utils.GetConfig()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.Home, 0755); err != nil {
		return fmt.Errorf("could not create pancake home directory %s: %w", cfg.Home, err)
	}
	fmt.Printf("Pancake home directory ready at %s\n", cfg.Home)

	if err := setupTools(); err != nil {
		fmt.Printf("Warning: tool setup skipped: %v\n", err)
	} else {
		fmt.Println("Tool setup completed.")
	}

	fmt.Println("Setup completed. Run 'pancake help' to see commands.")
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	initCmd := &cobra.Command{
		Use: "init",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initCommand(initForce); err != nil {
				fmt.Println("Error:", err)
				fmt.Println(utils.ConfigHintEditConfig)
				os.Exit(1)
			}
		},
	}
	initCmd.Flags().BoolVar(&initForce, "force", false, "Reset pancake.yml (backs up the old one)")

	rootCmd.AddCommand(
		&cobra.Command{
			Use:     "version",
			Aliases: []string{"v"},
			Run:     func(cmd *cobra.Command, args []string) { version() },
		},
		&cobra.Command{
			Use:     "edit config",
			Aliases: []string{"ec"},
			Run: func(cmd *cobra.Command, args []string) {
				if err := editConfig(); err != nil {
					fmt.Println("Error:", err)
					os.Exit(1)
				}
			},
		},
		initCmd,
	)
}
