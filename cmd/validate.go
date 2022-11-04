/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

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

type TerraformConfigValidatorInput struct {
	TerraformPlan             ParsedTerraformPlan `json:"terraform_plan"`
	Organization              string              `json:"organization"`
	Repository                string              `json:"repository"`
	TriggeringActionEventName string              `json:"event_name"`
	IssueNumber               int                 `json:"issue_number"`
	CommitSHA                 string              `json:"commit_sha"`
	Ref                       string              `json:"ref"`
}

const validationURL = "https://app.getdigraph.com/api/validate/terraform"

func parseTerraformPlanJSON(jsonFilePath string) (ParsedTerraformPlan, error) {
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		return ParsedTerraformPlan{}, fmt.Errorf("error %s", err.Error())
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

	defer jsonFile.Close()

	// cleanup by removing temp file that was written
	os.Remove(jsonFilePath)
	return parsedPlanChangesOnly, nil
}

func invokeDigraphValidateAPI(parsedTFPlan ParsedTerraformPlan, digraphAPIKey, organization, repository, eventName, ref, commitSHA string, issueNumber int) error {
	requestBody := TerraformConfigValidatorInput{
		TerraformPlan:             parsedTFPlan,
		Organization:              organization,
		Repository:                repository,
		Ref:                       ref,
		TriggeringActionEventName: eventName,
	}

	if issueNumber > 0 {
		requestBody.IssueNumber = issueNumber
	} else if len(commitSHA) > 0 {
		requestBody.CommitSHA = commitSHA
	} else {
		return errors.New("invalid input- must specify pull request or commit sha")
	}

	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, validationURL, bytes.NewReader(requestBytes))
	if err != nil {
		return err
	}

	req.Header.Set("X-Digraph-Secret-Key", digraphAPIKey)

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	_, err = client.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return err
	}

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
			digraphAPIKey, _ := cmd.Flags().GetString("digraphAPIKey")
			organization, _ := cmd.Flags().GetString("organization")
			repository, _ := cmd.Flags().GetString("repository")
			eventName, _ := cmd.Flags().GetString("eventName")
			ref, _ := cmd.Flags().GetString("ref")

			issueNumber, _ := cmd.Flags().GetInt("issueNumber")
			commitSHA, _ := cmd.Flags().GetString("commitSHA")

			execCmd := exec.Command("/bin/sh", "/scripts/get_remote_plan_json.sh", tfPlanOutput, terraformAPIKey)
			jsonPathBytes, err := execCmd.Output()
			if err != nil {
				return fmt.Errorf("error getting json plan %s", err.Error())
			}

			fmt.Print("Successfully got JSON plan")

			jsonFilePath := strings.TrimSpace(strings.Replace(string(jsonPathBytes), "\r\n", "", -1))

			parsedPlan, err := parseTerraformPlanJSON(jsonFilePath)
			if err != nil {
				return fmt.Errorf("error parsing JSON %s", err.Error())
			}

			err = invokeDigraphValidateAPI(parsedPlan, digraphAPIKey, organization, repository, eventName, ref, commitSHA, issueNumber)
			if err != nil {
				return fmt.Errorf("error calling API %s", err.Error())
			}
			return nil
		},
	}

	cmd.Flags().String("stdOutPlan", "", "StdOut for terraform plan command")
	_ = cmd.MarkFlagRequired("stdOutPlan")

	cmd.Flags().String("terraformAPIKey", "", "Terraform API Key")
	_ = cmd.MarkFlagRequired("terraformAPIKey")

	cmd.Flags().String("digraphAPIKey", "", "Digraph API Key")
	_ = cmd.MarkFlagRequired("digraphAPIKey")

	cmd.Flags().String("organization", "", "Github organization")
	cmd.Flags().String("repository", "", "Github repository")
	cmd.Flags().String("eventName", "", "Pull or Push")
	cmd.Flags().String("ref", "", "Branch ref")

	cmd.Flags().Int("issueNumber", 0, "Pull Request Number")
	cmd.Flags().String("commitSHA", "", "Commit SHA")

	return cmd
}
