/*
Copyright © 2021 Morgan Hein <work@morganhe.in>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"github.com/morganhein/envy/pkg/io"
	"github.com/morganhein/envy/pkg/manager"
	"github.com/spf13/cobra"
	"time"
)

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use:   "task [taskName]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cobra.CheckErr("need task name")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
		defer cancel()
		sh, err := io.CreateShell()
		cobra.CheckErr(err)
		mgr := manager.New(io.NewFilesystem(), sh)
		appConfig := manager.RunConfig{
			RecipeLocation: cfgFile,
			Operation:      manager.TASK,
			Sudo:           sudo,
			Verbose:        verbose,
			DryRun:         dryRun,
			ForceInstaller: "", //TODO (@morgan): add this to the cobra loading
		}
		err = mgr.Start(ctx, appConfig, manager.TASK, args[0])
		if err != nil {
			cobra.CheckErr(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
}
