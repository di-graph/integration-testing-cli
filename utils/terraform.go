package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/tidwall/gjson"
	"mvdan.cc/xurls/v2"
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

func ParseTerraformURL(tfPlanOutput string) (string, error) {
	rxStrict := xurls.Strict()
	tfUrl := rxStrict.FindString(tfPlanOutput)
	if len(tfUrl) == 0 {
		return "", errors.New("could not find url")
	}
	return tfUrl, nil
}

func ParseTerraformRunID(tfUrl string) (string, error) {
	runId := tfUrl[strings.LastIndex(tfUrl, "/")+1:]
	if len(runId) == len(tfUrl) {
		return "", errors.New("could not parse run id from url")
	}
	return runId, nil
}

func GetTerraformPlanID(runId string, terraformAPIKey string) (string, error) {

	// PLAN_ID=`curl -N -s \
	// --header "Authorization: Bearer $2" \
	// https://app.terraform.io/api/v2/runs/$RUN_ID | python3 -c "import sys, json; print(json.load(sys.stdin)['data']['relationships']['plan']['data']['id'])"`
	// echo $PLAN_ID

	terraformRunUrl := fmt.Sprintf("https://app.terraform.io/api/v2/runs/%s", runId)
	bearerHeader := fmt.Sprintf("Bearer %s", terraformAPIKey)

	req, err := http.NewRequest("GET", terraformRunUrl, nil)
	if err != nil {
		return "", err
	}

	// add authorization header to the req
	req.Header.Add("Authorization", bearerHeader)

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	planId := gjson.Get(string(body), "data.relationships.plan.data.id")
	if !planId.Exists() {
		return "", errors.New("could not fetch plan id from output")
	}
	return planId.String(), nil
}

func GetAndSaveTerraformPlanJSON(planId string, terraformAPIKey string) (string, error) {
	terraformPlanJSONFilePath := "/tmp/tf_plan.json"
	terraformRunUrl := fmt.Sprintf("https://app.terraform.io/api/v2/plans/%s/json-output-redacted", planId)
	bearerHeader := fmt.Sprintf("Bearer %s", terraformAPIKey)

	req, err := http.NewRequest("GET", terraformRunUrl, nil)
	if err != nil {
		return "", err
	}

	// add authorization header to the req
	req.Header.Add("Authorization", bearerHeader)
	req.Header.Add("Content-Type", "application/vnd.api+json")

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tempFile, err := os.Create(terraformPlanJSONFilePath)
	if err != nil {
		return "", err
	}

	defer tempFile.Close()
	_, err = tempFile.Write(body)
	if err != nil {
		return "", err
	}

	return terraformPlanJSONFilePath, nil
}

func FetchRemoteTerraformPlan(tfPlanOutput string, terraformAPIKey string) (string, error) {
	url, err := ParseTerraformURL(tfPlanOutput)
	if err != nil {
		return "", fmt.Errorf("error getting url %s", err.Error())
	}

	runId, err := ParseTerraformRunID(url)
	if err != nil {
		return "", fmt.Errorf("error getting run id %s", err.Error())
	}

	planId, err := GetTerraformPlanID(runId, terraformAPIKey)
	if err != nil {
		return "", fmt.Errorf("error getting plan id %s", err.Error())
	}

	jsonFilePath, err := GetAndSaveTerraformPlanJSON(planId, terraformAPIKey)
	if err != nil {
		return "", fmt.Errorf("error getting and saving terraform plan %s", err.Error())
	}

	return jsonFilePath, nil
}

func ParseTerraformPlanJSON(jsonFilePath string) (ParsedTerraformPlan, error) {
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

	return parsedPlanChangesOnly, nil
}
