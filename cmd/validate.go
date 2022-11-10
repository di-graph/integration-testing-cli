/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
	InvocationMode            string                    `json:"invocation_mode"`
}

const validationURL = "https://app.getdigraph.com/api/validate/terraform"

func invokeDigraphValidateAPI(parsedTFPlan utils.ParsedTerraformPlan, digraphAPIKey, mode, repository, ref, commitSHA string, issueNumber int) (string, error) {
	requestBody := TerraformConfigValidatorInput{
		TerraformPlan:  parsedTFPlan,
		Repository:     repository,
		Ref:            ref,
		InvocationMode: mode,
	}

	if mode == "ci/cd" {
		if issueNumber > 0 {
			requestBody.IssueNumber = issueNumber
			requestBody.TriggeringActionEventName = "pull_request"
		} else if len(commitSHA) > 0 {
			requestBody.CommitSHA = commitSHA
			requestBody.TriggeringActionEventName = "push"
		} else {
			return "", errors.New("invalid input- must specify pull request or commit sha")
		}

		if len(commitSHA) > 0 && len(ref) == 0 {
			return "", errors.New("must provide branch ref associated with commit")
		}

		if len(repository) == 0 {
			return "", errors.New("must provide repository flag")
		}
	}

	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, validationURL, bytes.NewReader(requestBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("X-Digraph-Secret-Key", digraphAPIKey)

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return "", err
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	return string(body), nil
}

func validate() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "validate",
		Short:     "Validate infra config changes",
		Long:      ``,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			tfPlanOutput, _ := cmd.Flags().GetString("raw-output-plan")
			jsonPathPlan, _ := cmd.Flags().GetString("json-path-plan")
			tfRawPath, _ := cmd.Flags().GetString("output-path-plan")

			terraformAPIKey, _ := cmd.Flags().GetString("terraform-api-key")
			digraphAPIKey, _ := cmd.Flags().GetString("api-key")
			repository, _ := cmd.Flags().GetString("repository")
			ref, _ := cmd.Flags().GetString("ref")

			issueNumber, _ := cmd.Flags().GetInt("issue-number")
			commitSHA, _ := cmd.Flags().GetString("commit-sha")

			mode, _ := cmd.Flags().GetString("mode")

			if len(digraphAPIKey) == 0 {
				err := godotenv.Load(".env")

				if err != nil {
					return fmt.Errorf("must specify api-key as argument or set it within a .env file")
				}

				digraphAPIKey = os.Getenv("api-key")
			}

			var jsonFilePath string
			var err error
			if len(tfPlanOutput) > 0 {
				if len(terraformAPIKey) == 0 {
					err := godotenv.Load(".env")

					if err != nil {
						return fmt.Errorf("must specify terraform-api-key as argument or set it within a .env file")
					}

					terraformAPIKey = os.Getenv("terraform-api-key")
				}

				jsonFilePath, err = utils.FetchRemoteTerraformPlan(tfPlanOutput, terraformAPIKey)
				if err != nil {
					return fmt.Errorf("error getting plan json %s", err.Error())
				}
			} else if len(jsonPathPlan) > 0 {
				jsonFilePath = jsonPathPlan
			} else if len(tfRawPath) > 0 {
				if len(terraformAPIKey) == 0 {
					err := godotenv.Load(".env")

					if err != nil {
						return fmt.Errorf("must specify terraform-api-key as argument or set it within a .env file")
					}

					terraformAPIKey = os.Getenv("terraform-api-key")
				}
				rawOutputFile, err := os.Open(tfRawPath)
				if err != nil {
					return fmt.Errorf("error %s", err.Error())
				}

				rawByteValue, _ := ioutil.ReadAll(rawOutputFile)
				jsonFilePath, err = utils.FetchRemoteTerraformPlan(string(rawByteValue), terraformAPIKey)
				if err != nil {
					return fmt.Errorf("error getting plan json %s", err.Error())
				}
			} else {
				return fmt.Errorf("must specify raw-output-plan or json-path-plan or output-path-plan")
			}

			parsedPlan, err := utils.ParseTerraformPlanJSON(jsonFilePath)
			if err != nil {
				return fmt.Errorf("error parsing JSON %s", err.Error())
			}

			output, err := invokeDigraphValidateAPI(parsedPlan, digraphAPIKey, mode, repository, ref, commitSHA, issueNumber)
			if err != nil {
				return fmt.Errorf("error calling API %s", err.Error())
			}

			if len(tfPlanOutput) > 0 {
				// cleanup by removing temp file that was written for terraform output case
				os.Remove(jsonFilePath)
			}
			if mode == "cli" {
				fmt.Printf("%s\n", output)
			}
			return nil
		},
	}

	cmd.Flags().String("raw-output-plan", "", "Terminal output from terraform plan command")
	cmd.Flags().String("output-path-plan", "", "Filepath for terminal output from terraform plan command")
	cmd.Flags().String("json-path-plan", "", "Filepath to terraform plan JSON file")

	cmd.Flags().String("terraform-api-key", "", "Terraform API Key")

	cmd.Flags().String("api-key", "", "Digraph API Key")

	cmd.Flags().String("repository", "", "Github repository")

	cmd.Flags().String("ref", "", "Branch ref")
	cmd.Flags().Int("issue-number", 0, "Pull Request Number")
	cmd.Flags().String("commit-sha", "", "Commit SHA")

	cmd.Flags().String("mode", "ci/cd", "Running mode- ci/cd or cli")

	return cmd
}
