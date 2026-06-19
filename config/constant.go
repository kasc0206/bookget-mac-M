package config

import (
	"os"
	"path/filepath"
)

const (
	Version              = "25.0701"
	CatalogVersionInfo   = "#版本=1.0" // 书签目录版本TXT
	defaultUserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36"
	defaultFileExtension = ".jpg"
)

func UserHomeDir() string {
	if os.PathSeparator == '\\' {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func BookgetHomeDir() string {
	home, err := os.UserHomeDir()
	if err == nil {
		// Unix-like: ~/bookget/path
		// Windows: ~\bookget\path
		configDir := filepath.Join(home, "bookget")
		if os.PathSeparator == '\\' { // Windows
			configDir = filepath.Join(home, "bookget")
		}
		homeDir := filepath.Join(configDir)
		if err := os.Mkdir(homeDir, 0755); err != nil && !os.IsExist(err) {
			return ""
		}
		return homeDir
	}
	return home
}

func CacheDir() string {
	return filepath.Join(BookgetHomeDir(), "cache")
}
