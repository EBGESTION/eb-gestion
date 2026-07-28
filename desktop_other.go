//go:build !windows

package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

// runDesktop permite validar el servidor fuera de Windows. La aplicación final
// usa WebView2 y se compila exclusivamente en el runner Windows de GitHub.
func runDesktop(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	if err := command.Start(); err != nil {
		return fmt.Errorf("abre manualmente %s: %w", url, err)
	}
	return command.Wait()
}
