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

	"github.com/a6h15hek/pancake/utils"
	"github.com/spf13/cobra"
)

/*
- This will have implementation of "pancake gpt <user_description_of_command>".
It will utilize the GPT models to understand the user's natural language input
in <user_description_of_command>, interpret it, create a corresponding command,
and execute it.
*/

var gptCmd = &cobra.Command{
	Use: "gpt",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.NotImplemented)
	},
}

func init() {
	rootCmd.AddCommand(gptCmd)
}
