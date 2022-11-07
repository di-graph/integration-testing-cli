/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// // multipleCmd represents the multiple command
// var multipleCmd = &cobra.Command{
// 	Use:   "multiple",
// 	Short: "A brief description of your command",
// 	Long: `A longer description that spans multiple lines and likely contains examples
// and usage of using your command. For example:

// Cobra is a CLI library for Go that empowers applications.
// This application is a tool to generate the needed files
// to quickly create a Cobra application.`,
// 	ValidArgs: []string{"--", "-"},
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		times, _ := cmd.Flags().GetString("times")
// 		fmt.Printf("multiple called with %s\n times", times)
// 	},
// }

// multipleCmd.Flags().String("commit", "", "Commit SHA to post comment on, mutually exclusive with pull-request")

// func init() {
// 	echoCmd.AddCommand(multipleCmd)

// 	// Here you will define your flags and configuration settings.

// 	// Cobra supports Persistent Flags which will work for this command
// 	// and all subcommands, e.g.:
// 	// multipleCmd.PersistentFlags().String("foo", "", "A help for foo")

// 	// Cobra supports local flags which will only run when this command
// 	// is called directly, e.g.:
// 	// multipleCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
// }

func multipleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multiple",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
		and usage of using your command. For example:
		
		Cobra is a CLI library for Go that empowers applications.
		This application is a tool to generate the needed files
		to quickly create a Cobra application.`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			times, _ := cmd.Flags().GetString("times")
			fmt.Printf("multiple called with %s times\n", times)

			return nil
		},
	}

	cmd.Flags().String("times", "", "Number of times the command was called")
	_ = cmd.MarkFlagRequired("times")

	return cmd
}
