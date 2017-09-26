/*
Copyright 2017 WALLIX

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

package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wallix/awless/database"
	"github.com/wallix/awless/logger"
	"github.com/wallix/awless/template"
)

var (
	deleteAllLogsFlag    bool
	deleteFromIdLogsFlag string
	limitLogCountFlag    int
	logsAsRawJSONFlag    bool
	logAsIDOnlyFlag      bool
	logsAsShortFlag      bool
)

func init() {
	RootCmd.AddCommand(logCmd)

	logCmd.Flags().BoolVar(&deleteAllLogsFlag, "delete-all", false, "Delete all logs from local db")
	logCmd.Flags().StringVar(&deleteFromIdLogsFlag, "delete", "", "Delete a specifc log entry given its id")
	logCmd.Flags().BoolVar(&logsAsRawJSONFlag, "raw-json", false, "Display logs as raw json")
	logCmd.Flags().BoolVar(&logsAsShortFlag, "short", false, "Display less detailled version of one or more template log")
	logCmd.Flags().BoolVar(&logAsIDOnlyFlag, "id-only", false, "Show only log template IDs (i.e. revert IDs)")
	logCmd.Flags().IntVarP(&limitLogCountFlag, "number", "n", 0, "Limit log output to the last n logs")
}

var logCmd = &cobra.Command{
	Use:               "log [REVERTID]",
	Short:             "Show all awless template actions against your cloud infrastructure",
	PersistentPreRun:  applyHooks(initLoggerHook, initAwlessEnvHook, firstInstallDoneHook),
	PersistentPostRun: applyHooks(verifyNewVersionHook, onVersionUpgrade),

	RunE: func(c *cobra.Command, args []string) error {
		var all []*database.LoadedTemplate

		if len(args) > 0 {
			exitOn(database.Execute(func(db *database.DB) error {
				single, err := db.GetLoadedTemplate(args[0])
				if err != nil {
					return err
				}
				all = append(all, single)
				return nil
			}))
			printAll(all, &logPrinter{os.Stdout})
			return nil
		}

		if deleteAllLogsFlag {
			exitOn(database.Execute(func(db *database.DB) error {
				return db.DeleteTemplates()
			}))
			return nil
		}

		if tid := deleteFromIdLogsFlag; tid != "" {
			exitOn(database.Execute(func(db *database.DB) error {
				return db.DeleteTemplate(tid)
			}))
			return nil
		}

		exitOn(database.Execute(func(db *database.DB) (dberr error) {
			all, dberr = db.ListTemplates()
			return
		}))

		printAll(all)
		return nil
	},
}

func printAll(all []*database.LoadedTemplate, printers ...template.Printer) {
	var printer template.Printer
	if len(printers) > 0 {
		printer = printers[0]
	} else {
		if logsAsRawJSONFlag {
			printer = template.NewJSONPrinter(os.Stdout)
		} else if logAsIDOnlyFlag {
			printer = &idOnlyPrinter{os.Stdout}
		} else {
			printer = &shortLogPrinter{os.Stdout}
		}
	}

	if limitLogCountFlag > 0 && limitLogCountFlag < len(all) {
		all = all[len(all)-limitLogCountFlag:]
	}

	for _, loaded := range all {
		if loaded.Err != nil {
			logger.Errorf("Template '%s' in error: %s", string(loaded.Key), loaded.Err)
			logger.Verbosef("Template raw content\n%s", loaded.Raw)
			fmt.Println()
			continue
		}

		if err := printer.Print(loaded.TplExec); err != nil {
			logger.Error(err.Error())
		}
		fmt.Println()
	}
}
