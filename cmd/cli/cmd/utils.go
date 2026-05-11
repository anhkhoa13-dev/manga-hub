package cmd

import (
	"os"
	"path/filepath"
)

func getTokenPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mangahub_token")
}

func saveToken(token string) error {
	return os.WriteFile(getTokenPath(), []byte(token), 0600)
}

func loadToken() string {
	data, err := os.ReadFile(getTokenPath())
	if err != nil {
		return ""
	}
	return string(data)
}