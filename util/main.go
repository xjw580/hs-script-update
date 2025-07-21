package util

import (
	"archive/zip"
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

func CopyDirectory(sourceDir, targetDir string, progress *walk.ProgressBar, appendLog func(log string), maxProgressValue int) error {
	fileCount, _ := CountFilesInDirectory(sourceDir)
	if fileCount <= 0 {
		fileCount = 1
	}
	step := max((maxProgressValue-progress.Value())/fileCount, 1)
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return nil
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
		//忽略文件
		if strings.HasSuffix(dstPath, ".flag") {
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
			return nil
		}

		if !info.IsDir() {
			fileCount++
		}

		return nil
	})
	if err != nil {
		return fileCount, err
	}

	return fileCount, nil
}

func DeleteLibFiles(folderPath, excludePath string, progress *walk.ProgressBar, appendLog func(log string), maxProgressValue int) error {
	step := max((maxProgressValue-progress.Value())/100, 1)
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".jar") || strings.HasSuffix(info.Name(), ".dll")) {
			if strings.Contains(path, excludePath) || strings.Contains(path, "plugin") {
				appendLog("跳过删除文件 " + path)
			} else {
				err := os.Remove(path)
				if err != nil {
					return err
				}
				progress.SetValue(min(maxProgressValue, progress.Value()+step))
				appendLog("删除文件 " + path)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func Unzip(zipFilePath string, destination string) error {
	// 打开 ZIP 文件
	zipFile, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// 创建目标目录
	os.MkdirAll(destination, os.ModePerm)

	// 遍历 ZIP 文件中的每个文件
	for _, file := range zipFile.File {
		// 构造目标文件的完整路径
		filePath := filepath.Join(destination, file.Name)

		// 检查是否是目录
		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		// 创建目标文件
		outFile, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		// 打开 ZIP 文件中的文件
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		// 复制内容到目标文件
		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}
	return nil
}
