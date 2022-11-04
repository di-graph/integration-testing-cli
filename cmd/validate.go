/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

type ResourceChangeDetails struct {
	Actions         []string    `json:"actions"`
	Before          interface{} `json:"before"`
	After           interface{} `json:"after"`
	AfterUnknown    interface{} `json:"after_unknown"`
	BeforeSensitive interface{} `json:"before_sensitive"`
	AfterSensitive  interface{} `json:"after_sensitive"`
}

type ResourceChange struct {
	Address      string                `json:"address"`
	Mode         string                `json:"mode"`
	Type         string                `json:"type"`
	Name         string                `json:"name"`
	ProviderName string                `json:"provider_name"`
	Change       ResourceChangeDetails `json:"change"`
}

type ParsedTerraformPlan struct {
	ResourceChanges []ResourceChange `json:"resource_changes"`
}

func parseTerraformPlanJSON(jsonFilePath string) error {
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		return fmt.Errorf("error %s", err.Error())
	}

	jsonByteValue, _ := ioutil.ReadAll(jsonFile)

	var parsedJSONPlan ParsedTerraformPlan

	json.Unmarshal(jsonByteValue, &parsedJSONPlan)

	var actualChanges []ResourceChange
	for _, resourceChange := range parsedJSONPlan.ResourceChanges {
		for _, action := range resourceChange.Change.Actions {
			if action != "no-op" {
				actualChanges = append(actualChanges, resourceChange)
			}
		}
	}

	var parsedPlanChangesOnly = ParsedTerraformPlan{
		ResourceChanges: actualChanges,
	}

	fmt.Printf("%v\n", parsedPlanChangesOnly)

	defer jsonFile.Close()

	// cleanup by removing temp file that was written
	os.Remove(jsonFilePath)
	return nil
}

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
