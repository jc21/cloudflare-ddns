package helper

import (
	"os/user"
	"strings"

	"github.com/jc21/cloudflare-ddns/internal/logger"
)

// GetFullFilename replaces wildcards in filenames
func GetFullFilename(filename string) string {
	usr, err := user.Current()
	if err != nil {
		logger.Error("%s", err.Error())
	}

	var strs []string
	strs = append(strs, usr.HomeDir)
	strs = append(strs, "/")

	return strings.ReplaceAll(filename, "~/", strings.Join(strs, ""))
}
