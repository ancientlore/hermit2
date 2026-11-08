package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ancientlore/hermit2/views"
)

// Options holds the persistent user options and state saved in options.toml.
type Options struct {
	SortBy        views.SortType      `toml:"sort_by"`
	SortDirection views.SortDirection `toml:"sort_direction"`
	Filter        string              `toml:"filter"`
	FilterHistory []string            `toml:"filter_history"`
	GotoHistory   []string            `toml:"goto_history"`
}

var activeOptionsPath string

// DefaultOptions returns an Options struct initialized with default settings.
func DefaultOptions() Options {
	return Options{
		SortBy:        views.SortByName,
		SortDirection: views.SortForward,
		Filter:        "",
		FilterHistory: nil,
		GotoHistory:   nil,
	}
}

// OptionsFile returns the full path to the options file in the user's config folder.
func OptionsFile() (string, error) {
	folder, err := ConfigFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(folder, OptionsFileName), nil
}

// InitOptions resolves the user's options file path, enables automatic saving to
// that path, and loads any existing options from disk.
func InitOptions() (Options, error) {
	path, err := OptionsFile()
	if err != nil {
		return DefaultOptions(), err
	}
	return InitOptionsAt(path)
}

// InitOptionsAt sets the active options file path to path (or disables automatic
// saving when path is empty) and loads options from that path.
func InitOptionsAt(path string) (Options, error) {
	activeOptionsPath = path
	if path == "" {
		return DefaultOptions(), nil
	}
	return LoadOptionsFrom(path)
}

// LoadOptions loads the user's options from the default options file.
// If the file does not exist yet, DefaultOptions is returned without error.
func LoadOptions() (Options, error) {
	path, err := OptionsFile()
	if err != nil {
		return DefaultOptions(), err
	}
	return LoadOptionsFrom(path)
}

// LoadOptionsFrom loads options from the specified TOML file path.
// If the file does not exist yet, DefaultOptions is returned without error.
func LoadOptionsFrom(path string) (Options, error) {
	opts := DefaultOptions()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return opts, nil
		}
		return opts, fmt.Errorf("cannot read options file %q: %w", path, err)
	}
	if _, err := toml.Decode(string(data), &opts); err != nil {
		return DefaultOptions(), fmt.Errorf("cannot parse options file %q: %w", path, err)
	}
	return opts, nil
}

// SaveOptions writes opts to the default options file in the user's config folder.
func SaveOptions(opts Options) error {
	path, err := OptionsFile()
	if err != nil {
		return err
	}
	return SaveOptionsTo(path, opts)
}

// SaveActiveOptions writes opts to the active options file configured by InitOptions
// or InitOptionsAt. If no active path is configured, it is a no-op.
func SaveActiveOptions(opts Options) error {
	if activeOptionsPath == "" {
		return nil
	}
	return SaveOptionsTo(activeOptionsPath, opts)
}

// SaveOptionsTo writes opts as TOML to the specified file path.
func SaveOptionsTo(path string, opts Options) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("cannot create options folder for %q: %w", path, err)
	}
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(opts); err != nil {
		return fmt.Errorf("cannot encode options: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("cannot write options file %q: %w", path, err)
	}
	return nil
}

