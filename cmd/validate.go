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
	"os"
	"time"

	"github.com/di-graph/integration-testing-cli/utils"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type TerraformConfigValidatorInput struct {
	TerraformPlan             utils.ParsedTerraformPlan `json:"terraform_plan"`
	Repository                string                    `json:"repository"`
	TriggeringActionEventName string                    `json:"event_name"`
	IssueNumber               int                       `json:"issue_number"`
	CommitSHA                 string                    `json:"commit_sha"`
	Ref                       string                    `json:"ref"`
}

const validationURL = "https://app.getdigraph.com/api/validate/terraform"

func invokeDigraphValidateAPI(parsedTFPlan utils.ParsedTerraformPlan, digraphAPIKey, repository, ref, commitSHA string, issueNumber int) error {
	requestBody := TerraformConfigValidatorInput{
		TerraformPlan: parsedTFPlan,
		Repository:    repository,
		Ref:           ref,
	}

	if issueNumber > 0 {
		requestBody.IssueNumber = issueNumber
		requestBody.TriggeringActionEventName = "pull_request"
	} else if len(commitSHA) > 0 {
		requestBody.CommitSHA = commitSHA
		requestBody.TriggeringActionEventName = "push"
	} else {
		return errors.New("invalid input- must specify pull request or commit sha")
	}

	if len(commitSHA) > 0 && len(ref) == 0 {
		return errors.New("must provide branch ref associated with commit")
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

func validate() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "validate",
		Short:     "Validate infra config changes",
		Long:      ``,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			tfPlanOutput, _ := cmd.Flags().GetString("stdOutPlan")
			jsonPathPlan, _ := cmd.Flags().GetString("jsonPathPlan")

			terraformAPIKey, _ := cmd.Flags().GetString("terraformAPIKey")
			digraphAPIKey, _ := cmd.Flags().GetString("digraphAPIKey")
			repository, _ := cmd.Flags().GetString("repository")
			ref, _ := cmd.Flags().GetString("ref")

			issueNumber, _ := cmd.Flags().GetInt("issueNumber")
			commitSHA, _ := cmd.Flags().GetString("commit-sha")

			if len(digraphAPIKey) == 0 {
				err := godotenv.Load(".env")

				if err != nil {
					return fmt.Errorf("must specify digraphAPIKey as argument or set it within a .env file")
				}

				digraphAPIKey = os.Getenv("digraphAPIKey")
			}

			if len(terraformAPIKey) == 0 {
				err := godotenv.Load(".env")

				if err != nil {
					return fmt.Errorf("must specify terraformAPIKey as argument or set it within a .env file")
				}

				terraformAPIKey = os.Getenv("terraformAPIKey")
			}

			var jsonFilePath string
			var err error
			if len(tfPlanOutput) > 0 {
				jsonFilePath, err = utils.FetchRemoteTerraformPlan(tfPlanOutput, terraformAPIKey)
				if err != nil {
					return fmt.Errorf("error getting plan json %s", err.Error())
				}
			} else if len(jsonPathPlan) > 0 {
				jsonFilePath = jsonPathPlan
			} else {
				return fmt.Errorf("must specify either stdOutPlan or jsonPathPlan")
			}

			parsedPlan, err := utils.ParseTerraformPlanJSON(jsonFilePath)
			if err != nil {
				return fmt.Errorf("error parsing JSON %s", err.Error())
			}

			err = invokeDigraphValidateAPI(parsedPlan, digraphAPIKey, repository, ref, commitSHA, issueNumber)
			if err != nil {
				return fmt.Errorf("error calling API %s", err.Error())
			}

			if len(tfPlanOutput) > 0 {
				// cleanup by removing temp file that was written for terraform output case
				os.Remove(jsonFilePath)
			}
			return nil
		},
	}

	cmd.Flags().String("stdOutPlan", "", "Terminal output from terraform plan command")
	cmd.Flags().String("jsonPathPlan", "", "Filepath to terraform plan JSON file")

	cmd.Flags().String("terraformAPIKey", "", "Terraform API Key")

	cmd.Flags().String("digraphAPIKey", "", "Digraph API Key")

	cmd.Flags().String("repository", "", "Github repository")
	_ = cmd.MarkFlagRequired("repository")

	cmd.Flags().String("ref", "", "Branch ref")
	cmd.Flags().Int("issueNumber", 0, "Pull Request Number")
	cmd.Flags().String("commit-sha", "", "Commit SHA")

	return cmd
}
