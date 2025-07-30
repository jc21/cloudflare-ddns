package model

// ArgConfig is the settings for passing arguments to the command
type ArgConfig struct {
	ConfigFile string `arg:"-c" help:"Config File to use (default: ~/.config/cloudflare-ddns.json)"`
	StateFile  string `arg:"-t" help:"State File to use (default: ~/.config/cloudflare-ddns-state.json)"`
	Setup      bool   `arg:"-s" help:"Setup wizard"`
	Force      bool   `arg:"-f" help:"Force update Cloudflare even if IP hasn't changed"`
	Quiet      bool   `arg:"-q" help:"Only print errors"`
	Verbose    bool   `arg:"-v" help:"Print a lot more info"`
}

// Description returns a simple description of the command
func (ArgConfig) Description() string {
	return "Update Cloudflare DNS record with your current IP address"
}
