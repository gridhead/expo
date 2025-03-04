package main

import (
	"errors"
	"fmt"
	"github.com/gridhead/expo/expo/base"
	"github.com/gridhead/expo/expo/item"
	"github.com/gridhead/expo/expo/task"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

func main() {
	var expt error

	base.SetLogger()

	var repodata item.RepoData
	var srut, drut, srce, dest, pkey, fkey, pusr, fusr string
	var rootTask = &cobra.Command{
		Use:   "expo",
		Short: "Forgejo Support for Pagure Exporter",
		PersistentPreRunE: func(comd *cobra.Command, args []string) error {
			if srce == "" {
				expt = errors.New("Source namespace not provided")
			}
			if dest == "" {
				expt = errors.New("Destination namespace not provided")
			}
			if pkey == "" {
				expt = errors.New("API key for Pagure not provided")
			}
			if fkey == "" {
				expt = errors.New("API key for Forgejo not provided")
			}
			if pusr == "" {
				expt = errors.New("Username for Pagure not provided")
			}
			if fusr == "" {
				expt = errors.New("Username for Forgejo not provided")
			}
			return expt
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			expt = errors.New("No subcommand executed")
			return expt
		},
	}
	rootTask.PersistentFlags().StringVarP(&srut, "srut", "x", "pagure.io", "Root of the source namespace")
	rootTask.PersistentFlags().StringVarP(&drut, "drut", "y", "test.gridhead.net", "Root of the destination namespace")
	rootTask.PersistentFlags().StringVarP(&srce, "srce", "s", "", "Namespace of source project for importing assets from")
	rootTask.PersistentFlags().StringVarP(&dest, "dest", "d", "", "Namespace of destination project for exporting assets to")
	rootTask.PersistentFlags().StringVarP(&pkey, "pkey", "p", "", "API key from Pagure for accessing the source namespace")
	rootTask.PersistentFlags().StringVarP(&fkey, "fkey", "f", "", "API key from Forgejo accessing the destination namespace")
	rootTask.PersistentFlags().StringVarP(&pusr, "pusr", "i", "", "Username of the account that owns the Pagure API key")
	rootTask.PersistentFlags().StringVarP(&fusr, "fusr", "j", "", "Username of the account that owns the Forgejo API key")
	rootTask.MarkPersistentFlagRequired("srce")
	rootTask.MarkPersistentFlagRequired("dest")
	rootTask.MarkPersistentFlagRequired("pkey")
	rootTask.MarkPersistentFlagRequired("fkey")
	rootTask.MarkPersistentFlagRequired("pusr")
	rootTask.MarkPersistentFlagRequired("fusr")

	var tktsdata item.TktsTaskData
	var status, ranges, choice string
	var comments, labels, commit, secret bool
	var rmin, rmax int
	var list []int
	var tktsTask = &cobra.Command{
		Use:   "tkts",
		Short: "Initialize transfer of issue tickets",
		Long:  "Initialize transfer of issue tickets",
		Args:  cobra.MinimumNArgs(0),
		PersistentPreRunE: func(comd *cobra.Command, args []string) error {
			var iner error
			if iner = task.ValidateStatusChoice(status); iner != nil {
				expt = iner
			}
			if ranges != "" {
				rmin, rmax, iner = task.ValidateRanges(ranges)
				if iner != nil {
					expt = iner
				}
			}
			if choice != "" {
				list, iner = task.ValidateChoice(choice)
				if iner != nil {
					expt = iner
				}
			}
			return expt
		},
		Run: func(comd *cobra.Command, args []string) {
			repodata = item.RepoData{
				RootSrce:     srut,
				NameSrce:     srce,
				RootDest:     drut,
				NameDest:     dest,
				UsernameSrce: pusr,
				UsernameDest: fusr,
				PasswordSrce: pkey,
				PasswordDest: fkey,
			}
			tktsdata = item.TktsTaskData{
				PerPageQuantity: 100,
				Ranges:          item.IssueTicketRanges{Min: rmin, Max: rmax},
				Choice:          list,
				Status:          task.ActionStatusChoice(status),
				WithComments:    comments,
				WithLabels:      labels,
				WithSecret:      secret,
				WithStatus:      commit,
				Retries:         4,
			}
			slog.Log(nil, slog.LevelInfo, fmt.Sprintf("%s", repodata))
			slog.Log(nil, slog.LevelInfo, fmt.Sprintf("%s", tktsdata))

			var here bool
			here, expt = task.VerifyProjects(&repodata)
			if !here {
				slog.Log(nil, slog.LevelError, fmt.Sprintf("✗ Either source namespace or destination namespace was not found"))
				os.Exit(1)
			} else {
				here, expt = task.FetchTransferQuantity(&repodata, &tktsdata)
				if expt != nil || here == false {
					slog.Log(nil, slog.LevelError, fmt.Sprintf("✗ Error occured. %s", expt.Error()))
					os.Exit(2)
				}
				os.Exit(0)
			}
		},
	}
	tktsTask.Flags().StringVarP(&status, "status", "u", "OPEN", "Extract issue tickets of the mentioned status")
	tktsTask.Flags().StringVarP(&ranges, "ranges", "r", "", "Extract issue tickets of the mentioned ranges")
	tktsTask.Flags().StringVarP(&choice, "choice", "e", "", "Extract issue tickets of the mentioned choice")
	tktsTask.Flags().BoolVarP(&comments, "comments", "c", false, "Transfer all the associated comments")
	tktsTask.Flags().BoolVarP(&commit, "commit", "a", false, "Assert issue ticket states as they were")
	tktsTask.Flags().BoolVarP(&labels, "labels", "l", false, "Migrate all the associated labels")
	tktsTask.Flags().BoolVarP(&secret, "secret", "t", false, "Confirm issue ticket privacy as they were")

	rootTask.AddCommand(tktsTask)

	if expt = rootTask.Execute(); expt != nil {
		os.Exit(1)
	}

	//os.Exit(0)
}
