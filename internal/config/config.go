package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jc21/cloudflare-ddns/internal/helper"
	"github.com/jc21/cloudflare-ddns/internal/logger"
	"github.com/jc21/cloudflare-ddns/internal/model"

	"github.com/AlecAivazis/survey/v2"
	"github.com/JeremyLoy/config"
	"github.com/alexflint/go-arg"
)

// Populated at build time using ldflags
var appArguments model.ArgConfig

const defaultConfigFile = "~/.config/cloudflare-ddns.json"

// GetConfig returns the ArgConfig
func GetConfig() model.ArgConfig {
	// nolint: gosec, errcheck
	config.FromEnv().To(&appArguments)
	arg.MustParse(&appArguments)

	return appArguments
}

// SetupConfig will ask for setup questions
func SetupConfig() {
	fmt.Println(`Refer to this guide to find your Account and Zone IDs:
  https://developers.cloudflare.com/fundamentals/account/find-account-and-zone-ids/`)

	// the questions to ask
	var questions = []*survey.Question{
		{
			Name:     "api_key",
			Prompt:   &survey.Input{Message: "Cloudflare API Key:"},
			Validate: survey.Required,
		},
		{
			Name:     "zone_id",
			Prompt:   &survey.Input{Message: "Zone ID:"},
			Validate: survey.Required,
		},
		{
			Name:     "dns_record",
			Prompt:   &survey.Input{Message: "DNS Record:"},
			Validate: survey.Required,
		},
		{
			Name: "protocols",
			Prompt: &survey.Select{
				Message: "Which IP Protocals do you update?",
				Options: []string{"IPv4 Only", "IPv6 Only ", "Both"},
			},
			Validate: survey.Required,
		},
		{
			Name:   "pushover_user_token",
			Prompt: &survey.Input{Message: "Pushover User Token: (leave blank to disable)"},
		},
	}

	// perform the questions
	var answers model.CloudflareConfig
	err := survey.Ask(questions, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	logger.Trace("Answers: %+v", answers)

	writeErr := answers.Write(getConfigFilename())
	if writeErr != nil {
		logger.Error("Could not write configuration: %v", writeErr.Error())
		os.Exit(1)
	}
}

func getConfigFilename() string {
	argConfig := GetConfig()
	if argConfig.ConfigFile != "" {
		return argConfig.ConfigFile
	}

	return helper.GetFullFilename(defaultConfigFile)
}

// GetCloudflareConfig returns the configuration as read from a file
func GetCloudflareConfig() model.CloudflareConfig {
	var cfg model.CloudflareConfig
	filename := getConfigFilename()

	// Make sure file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		logger.Error("Configuration not found, run again with -s")
		os.Exit(1)
	}

	// nolint: gosec
	jsonFile, err := os.Open(filename)
	if err != nil {
		logger.Error("Configuration could not be opened: %v", err.Error())
		os.Exit(1)
	}

	// nolint: gosec, errcheck
	defer jsonFile.Close()

	contents, readErr := io.ReadAll(jsonFile)
	if readErr != nil {
		logger.Error("Configuration file could not be read: %v", readErr.Error())
		// nolint: gocritic
		os.Exit(1)
	}

	unmarshalErr := json.Unmarshal(contents, &cfg)
	if unmarshalErr != nil {
		logger.Error("Configuration file looks damaged, run again with -s")
		os.Exit(1)
	}

	return cfg
}
