package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/jc21/cloudflare-ddns/internal/helper"
	"github.com/jc21/cloudflare-ddns/internal/logger"
	"github.com/jc21/cloudflare-ddns/internal/model"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	externalip "github.com/glendc/go-external-ip"
	"github.com/gregdel/pushover"
)

const defaultStateFile = "~/.config/cloudflare-ddns-state.json"

// Process will update the ip address with cloudflare, if forced or changed
func Process(argConfig model.ArgConfig, cfg model.CloudflareConfig) {
	// Determine the current public ip
	consensus := externalip.DefaultConsensus(nil, nil)
	state := GetState(argConfig)
	logger.Trace("STATE: %+v", state)

	// Apply state
	state.DNSRecord = cfg.DNSRecord
	state.LastUpdateTime = time.Now()

	if cfg.DNSRecord != state.DNSRecord {
		argConfig.Force = true
	}

	cfg.Protocols = strings.Trim(cfg.Protocols, " ")
	hasError := false

	var ipv4 net.IP
	var ipv6 net.IP
	var err error

	if cfg.IPv4Enabled() {
		// nolint: errcheck, gosec
		consensus.UseIPProtocol(4)
		ipv4, err = consensus.ExternalIP()
		if err != nil {
			logger.Error("%s", err.Error())
			hasError = true
		}
	}

	if cfg.IPv6Enabled() {
		// nolint: errcheck, gosec
		consensus.UseIPProtocol(6)
		ipv6, err = consensus.ExternalIP()
		if err != nil {
			logger.Error("%s", err.Error())
			hasError = true
		}
	}

	if hasError {
		os.Exit(1)
		return
	}

	if err := updateIPProtocol(ipv4, ipv6, state, argConfig, cfg); err != nil {
		logger.Error("Could not update Cloudflare: %v", err.Error())
		os.Exit(1)
	}
}

// updateIPProtocol ...
func updateIPProtocol(
	ipv4,
	ipv6 net.IP,
	state model.CloudflareState,
	argConfig model.ArgConfig,
	cfg model.CloudflareConfig,
) error {
	ipv4Str := ""
	ipv6Str := ""
	changedIPs := make([]string, 0)

	if cfg.IPv4Enabled() && ipv4 != nil && (ipv4.String() != state.LastIPv4 || argConfig.Force) {
		ipv4Str = ipv4.String()
		logger.Info("Updating IPv4 to %v", ipv4Str)
		changedIPs = append(changedIPs, ipv4Str)
	}
	if cfg.IPv6Enabled() && ipv6 != nil && (ipv6.String() != state.LastIPv6 || argConfig.Force) {
		ipv6Str = ipv6.String()
		logger.Info("Updating IPv6 to %v", ipv6Str)
		changedIPs = append(changedIPs, ipv6Str)
	}

	if ipv4Str != "" || ipv6Str != "" {
		if err := updateIP(cfg, ipv4Str, ipv6Str); err != nil {
			return err
		}
		if ipv4Str != "" {
			state.LastIPv4 = ipv4Str
		}
		if ipv6Str != "" {
			state.LastIPv6 = ipv6Str
		}
		if err := state.Write(getStateFilename(argConfig)); err != nil {
			return fmt.Errorf("could not write state file: %v", err.Error())
		}

		if cfg.PushoverUserToken != "" {
			pushoverApp := pushover.New("a4dhut1a7waegz6p2xh7enzegjedgo")
			recipient := pushover.NewRecipient(cfg.PushoverUserToken)

			message := &pushover.Message{
				Message:    fmt.Sprintf("For %v", cfg.DNSRecord),
				Title:      fmt.Sprintf("IP updated to %v", strings.Join(changedIPs, ", ")),
				Priority:   0,
				URL:        "",
				URLTitle:   "",
				Timestamp:  time.Now().Unix(),
				Retry:      60 * time.Second,
				Expire:     time.Hour,
				DeviceName: "",
				Sound:      "",
			}

			// Send the message to the recipient
			_, err := pushoverApp.SendMessage(message, recipient)
			if err != nil {
				logger.Error("%s", err.Error())
			} else {
				logger.Info("Pushover Notification Sent OK")
			}
		}
		return nil
	}

	logger.Info("IP hasn't changed, not updating Cloudflare")
	return nil
}

func updateIP(cfg model.CloudflareConfig, ipv4, ipv6 string) error {
	client := cloudflare.NewClient(option.WithAPIToken(cfg.APIKey))

	// Find the record that matches the DNSRecord and the recordType
	params := dns.RecordListParams{
		ZoneID: cloudflare.F(cfg.ZoneID),
		Match:  cloudflare.F(dns.RecordListParamsMatchAll),
		Name: cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.F(cfg.DNSRecord),
		}),
	}
	res, err := client.DNS.Records.List(context.Background(), params)
	if err != nil {
		return err
	}

	if len(res.Result) == 0 {
		return fmt.Errorf(
			"no DNS record found for %s in Zone %s. make sure they exist before attempting to update",
			cfg.DNSRecord,
			cfg.ZoneID,
		)
	}

	ipv4Updated := false
	ipv6Updated := false

	// Iterate through the records and update them if they match the DNSRecord
	for _, record := range res.Result {
		if record.Name != cfg.DNSRecord {
			// Just in case the record name doesn't match, skip it
			continue
		}
		if record.Type == "A" && ipv4 != "" {
			// Update IPv4 record
			if _, err := client.DNS.Records.Update(
				context.Background(),
				record.ID,
				dns.RecordUpdateParams{
					ZoneID: cloudflare.F(cfg.ZoneID),
					Body: dns.ARecordParam{
						Name:    cloudflare.F(record.Name),
						TTL:     cloudflare.F(record.TTL),
						Type:    cloudflare.F(dns.ARecordType(record.Type)),
						Content: cloudflare.F(ipv4),
						Comment: cloudflare.F("Updated by cloudflare-ddns"),
						Proxied: cloudflare.F(record.Proxied),
					},
				},
			); err != nil {
				return fmt.Errorf("could not update IPv4 record: %v", err.Error())
			}
			ipv4Updated = true
		}
		if record.Type == "AAAA" && ipv6 != "" {
			// Update IPv4 record
			if _, err := client.DNS.Records.Update(
				context.Background(),
				record.ID,
				dns.RecordUpdateParams{
					ZoneID: cloudflare.F(cfg.ZoneID),
					Body: dns.AAAARecordParam{
						Name:    cloudflare.F(record.Name),
						TTL:     cloudflare.F(record.TTL),
						Type:    cloudflare.F(dns.AAAARecordType(record.Type)),
						Content: cloudflare.F(ipv6),
						Comment: cloudflare.F("Updated by cloudflare-ddns"),
						Proxied: cloudflare.F(record.Proxied),
					},
				},
			); err != nil {
				return fmt.Errorf("could not update IPv6 record: %v", err.Error())
			}
			ipv6Updated = true
		}
	}

	if !ipv4Updated && ipv4 != "" {
		logger.Error("Could not find IPv4 (A) record for %s in Zone %s", cfg.DNSRecord, cfg.ZoneID)
	}
	if !ipv6Updated && ipv6 != "" {
		logger.Error("Could not find IPv6 (AAAA) record for %s in Zone %s", cfg.DNSRecord, cfg.ZoneID)
	}

	return nil
}

func getStateFilename(argConfig model.ArgConfig) string {
	if argConfig.StateFile != "" {
		return argConfig.StateFile
	}

	return helper.GetFullFilename(defaultStateFile)
}

// GetState returns the configuration as read from a file
func GetState(argConfig model.ArgConfig) model.CloudflareState {
	var state model.CloudflareState
	filename := getStateFilename(argConfig)

	// nolint: gosec
	jsonFile, err := os.Open(filename)
	if err == nil {
		// nolint: errcheck
		defer jsonFile.Close()
		contents, readErr := io.ReadAll(jsonFile)
		if readErr == nil {
			err := json.Unmarshal(contents, &state)
			if err != nil {
				logger.Error("State file looks damaged, run again with -s")
			}
		}
	}

	return state
}
