package main

import (
	"Hearthstone-Script-update/gui"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	tempDir := os.TempDir()
	executable, _ := os.Executable()
	if filepath.Clean(filepath.Dir(executable)) == filepath.Clean(tempDir) || len(os.Args) == 1 {
		gui.ShowWindow()
	} else {
		executable, _ := os.Executable()
		_ = exec.Command("cmd", "/c", "copy", "/Y", executable, tempDir).Run()
		execName := filepath.Base(executable)
		time.Sleep(200 * time.Millisecond)
		cmd := exec.Command("cmd", "/c", "start", filepath.Join(tempDir, execName))
		cmd.Args = append(cmd.Args, os.Args[1:]...)
		_ = cmd.Run()
	}
}
