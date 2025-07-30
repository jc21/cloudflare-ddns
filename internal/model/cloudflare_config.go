package model

import (
	"encoding/json"
	"os"
	"path"
)

// CloudflareConfig is the settings that are saved for use in updating
type CloudflareConfig struct {
	ZoneID            string `survey:"zone_id"`
	DNSRecord         string `survey:"dns_record"`
	APIKey            string `survey:"api_key"`
	Protocols         string `survey:"protocols"`
	PushoverUserToken string `survey:"pushover_user_token"`
}

func (c *CloudflareConfig) Write(filename string) error {
	content, _ := json.MarshalIndent(c, "", " ")

	// Make sure the ".config" folder exists
	folder := path.Dir(filename)
	// nolint: gosec
	dirErr := os.MkdirAll(folder, os.ModePerm)
	if dirErr != nil {
		return dirErr
	}

	err := os.WriteFile(filename, content, 0600)
	if err != nil {
		return err
	}

	return nil
}

// IPv4Enabled ...
func (c *CloudflareConfig) IPv4Enabled() bool {
	return c.Protocols == "IPv4 Only" || c.Protocols == "Both" || c.Protocols == ""
}

// IPv6Enabled ...
func (c *CloudflareConfig) IPv6Enabled() bool {
	return c.Protocols == "IPv6 Only" || c.Protocols == "Both"
}
