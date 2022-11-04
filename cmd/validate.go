/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

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

func validateTF() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "validate-terraform",
		Short:     "Validate terraform changes",
		Long:      ``,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			tfPlanOutput, _ := cmd.Flags().GetString("stdOutPlan")
			terraformAPIKey, _ := cmd.Flags().GetString("terraformAPIKey")
			// digraphAPIKey, _ := cmd.Flags().GetString("digraphAPIKey")

			execCmd := exec.Command("/bin/sh", "./scripts/get_remote_plan_json.sh", tfPlanOutput, terraformAPIKey)
			jsonPathBytes, err := execCmd.Output()
			if err != nil {
				return fmt.Errorf("error getting json plan %s", err.Error())
			}

			jsonFilePath := strings.TrimSpace(strings.Replace(string(jsonPathBytes), "\r\n", "", -1))

			err = parseTerraformPlanJSON(jsonFilePath)
			if err != nil {
				return fmt.Errorf("error parsing JSON %s", err.Error())
			}

			return nil
		},
	}

	cmd.Flags().String("stdOutPlan", "", "StdOut for terraform plan command")
	_ = cmd.MarkFlagRequired("stdOutPlan")

	cmd.Flags().String("terraformAPIKey", "", "Terraform API Key")
	_ = cmd.MarkFlagRequired("terraformAPIKey")

	cmd.Flags().String("digraphAPIKey", "", "Digraph API Key")
	// _ = cmd.MarkFlagRequired("digraphAPIKey")

	return cmd
}

func parseTerraformPlanJSON(jsonFilePath string) error {
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		return fmt.Errorf("error %s", err.Error())
	}

	jsonByteValue, _ := ioutil.ReadAll(jsonFile)
	fmt.Print(string(jsonByteValue))

	defer jsonFile.Close()
	os.Remove(jsonFilePath)
	return nil
}
