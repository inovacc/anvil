package application

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	// AppName is the application name used for directories and identification
	AppName = "profile"

	// AppExeName is the executable name (without extension)
	AppExeName = "profile"

	// AppExeNameWindows is the executable name on Windows
	AppExeNameWindows = "profile.exe"
)

var (
	once   sync.Once
	appDir string
	errDir error
)

// GetApplicationDirectory returns the profile configuration directory path.
// Linux: ~/.config/profile (via os.UserConfigDir)
// Windows: C:\Users\{username}\AppData\Local\profile (via os.UserCacheDir)
func GetApplicationDirectory() (string, error) {
	once.Do(lazyLoad)

	if errDir != nil {
		return "", errDir
	}

	return appDir, errDir
}

func lazyLoad() {
	var (
		baseDir string
		err     error
	)

	switch runtime.GOOS {
	case "windows":
		// Windows: use AppData\Local (via UserCacheDir)
		baseDir, err = os.UserCacheDir()
	default:
		// Linux/others: use ~/.config (via UserConfigDir)
		baseDir, err = os.UserConfigDir()
	}

	if err != nil {
		errDir = fmt.Errorf("failed to get config directory: %w", err)
	}

	appDir = filepath.Join(baseDir, AppName)
}
