//go:build windows

package config

import "os"

const (
	configFolderName = "Hermit"
	ConfigFileName   = "hermit.toml"
)

// HomeFolder returns the location of the user's home folder
// from the $USERPROFILE environment variable.
func HomeFolder() string {
	return os.Getenv("USERPROFILE")
}

// ConfigFolder returns the location of the hermit configuration folder,
// which is $APPDATA\Hermit.
func ConfigFolder() (string, error) {
	f := filepath.Join(os.Getenv("APPDATA"), configFolderName)
	s, err := os.Stat(f)
	// Found it; let's make sure it is a folder
	if err == nil {
		if !s.IsDir() {
			return "", fmt.Errorf("config folder is a regular file: %q", f)
		}
		return f, nil
	}
	// Doesn't exist; let's try to create it
	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(f, 0655)
		if err != nil {
			return "", fmt.Errorf("cannot create config folder %q: %w", f, err)
		}
		return f, nil
	}
	// Something else is wrong
	return "", err
}
