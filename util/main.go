package util

import (
	"github.com/lxn/walk"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func KillProcess(pid string) error {
	cmd := exec.Command("taskkill", "/PID", pid, "/F")
	return cmd.Run()
}

func copyFile(source, target string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	return os.Chmod(target, info.Mode())
}

func CopyDirectory(sourceDir, targetDir string, progress *walk.ProgressBar, appendLog func(log string)) error {
	fileCount, _ := CountFilesInDirectory(sourceDir)
	step := (progress.MaxValue() - progress.Value() - 5) / fileCount
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(targetDir, relPath)

		dstDirPath := filepath.Dir(dstPath)
		if _, err := os.Stat(dstDirPath); os.IsNotExist(err) {
			if err := os.MkdirAll(dstDirPath, os.ModePerm); err != nil {
				appendLog("创建文件夹失败 " + dstDirPath)
				return nil
			}
			appendLog("创建文件夹 " + dstDirPath)
		}
		if strings.HasSuffix(dstPath, "update.exe") {
			appendLog("跳过复制文件 " + path)
		} else {
			if err := copyFile(path, dstPath); err != nil {
				appendLog("复制文件失败 " + path)
			} else {
				appendLog("复制文件 " + path)
			}
		}
		progress.SetValue(progress.Value() + step)

		return nil
	})
}

func CountFilesInDirectory(dirPath string) (int, error) {
	fileCount := 0

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fileCount++
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return fileCount, nil
}
