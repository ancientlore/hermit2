//go:build windows

package config

import "os"

func Shell() string {
	shell := os.Getenv("HERMIT_SHELL")
	if shell == "" {
		shell = os.Getenv("SHELL")
	}
	if shell == "" {
		shell = os.Getenv("COMSPEC")
	}
	if shell == "" {
		shell = "cmd.exe"
	}
	return shell
}
