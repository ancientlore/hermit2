//go:build !windows && !linux && !darwin

package browser

import "fmt"

func sysInfo(s any) []string {
	return []string{fmt.Sprintf("%#v", s)}
}

