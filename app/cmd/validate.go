/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/di-graph/integration-testing-cli/utils"
	"github.com/spf13/cobra"
)

type TerraformConfigValidatorInput struct {
	TerraformPlan             utils.ParsedTerraformPlan `json:"terraform_plan"`
	Organization              string                    `json:"organization"`
	Repository                string                    `json:"repository"`
	TriggeringActionEventName string                    `json:"event_name"`
	IssueNumber               int                       `json:"issue_number"`
	CommitSHA                 string                    `json:"commit_sha"`
	Ref                       string                    `json:"ref"`
}

const validationURL = "https://app.getdigraph.com/api/validate/terraform"

func invokeDigraphValidateAPI(parsedTFPlan utils.ParsedTerraformPlan, digraphAPIKey, organization, repository, eventName, ref, commitSHA string, issueNumber int) error {
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

			jsonFilePath, err := utils.FetchRemoteTerraformPlan(tfPlanOutput, terraformAPIKey)
			if err != nil {
				return fmt.Errorf("error getting plan json %s", err.Error())
			}
			parsedPlan, err := utils.ParseTerraformPlanJSON(jsonFilePath)
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
