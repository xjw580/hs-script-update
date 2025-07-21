package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// FileExists 检查文件是否存在
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// KillProcess 终止指定PID的进程
func KillProcess(pid string) error {
	cmd := exec.Command("taskkill", "/PID", pid, "/F")
	return cmd.Run()
}

// CopyFile 复制单个文件
func CopyFile(source, target string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// 确保目标目录存在
	targetDir := filepath.Dir(target)
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return err
	}

	targetFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	// 复制文件内容
	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return err
	}

	// 复制文件权限
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	return os.Chmod(target, sourceInfo.Mode())
}

// CopyDirectory 复制整个目录
func CopyDirectory(sourceDir, targetDir string, logFunc func(string)) error {
	return CopyDirectoryWithExcludes(sourceDir, targetDir, nil, logFunc)
}

// CopyDirectoryWithExcludes 复制目录并排除指定文件
func CopyDirectoryWithExcludes(sourceDir, targetDir string, excludeFiles []string, logFunc func(string)) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, relPath)

		// 检查是否需要排除
		if shouldExcludeFile(info.Name(), excludeFiles) {
			logFunc("跳过复制文件: " + path + " (排除列表)")
			return nil
		}

		// 跳过特定文件
		if strings.HasSuffix(targetPath, ".flag") {
			logFunc("跳过复制文件: " + path + " (.flag文件)")
			return nil
		}

		// 复制文件
		if err := CopyFile(path, targetPath); err != nil {
			logFunc(fmt.Sprintf("复制文件失败 %s: %v", path, err))
			return err
		}

		logFunc("复制文件: " + path)
		return nil
	})
}

// shouldExcludeFile 检查文件是否应该被排除
func shouldExcludeFile(filename string, excludeFiles []string) bool {
	for _, exclude := range excludeFiles {
		if strings.EqualFold(filename, exclude) {
			return true
		}
	}
	return false
}

// CountFilesInDirectory 统计目录中的文件数量
func CountFilesInDirectory(dirPath string) (int, error) {
	fileCount := 0
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续处理
		}
		if !info.IsDir() {
			fileCount++
		}
		return nil
	})
	return fileCount, err
}

// DeleteLibFiles 删除库文件
func DeleteLibFiles(folderPath, excludePath string, logFunc func(string)) error {
	return filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理文件
		if info.IsDir() {
			return nil
		}

		// 检查是否是库文件
		if !isLibraryFile(info.Name()) {
			return nil
		}

		// 检查是否需要排除
		if shouldExclude(path, excludePath) {
			logFunc("跳过删除文件: " + path)
			return nil
		}

		// 删除文件
		if err := os.Remove(path); err != nil {
			logFunc(fmt.Sprintf("删除文件失败 %s: %v", path, err))
			return err
		}

		logFunc("删除文件: " + path)
		return nil
	})
}

// isLibraryFile 检查是否是库文件
func isLibraryFile(filename string) bool {
	return strings.HasSuffix(filename, ".jar") || strings.HasSuffix(filename, ".dll")
}

// shouldExclude 检查是否应该排除文件
func shouldExclude(path, excludePath string) bool {
	return strings.Contains(path, excludePath) || strings.Contains(path, "plugin")
}

// UnzipFile 解压ZIP文件
func UnzipFile(zipFilePath, destination string) error {
	// 打开ZIP文件
	zipReader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	// 创建目标目录
	if err := os.MkdirAll(destination, os.ModePerm); err != nil {
		return err
	}

	// 解压每个文件
	for _, file := range zipReader.File {
		filePath := filepath.Join(destination, file.Name)

		// 安全检查：防止路径遍历攻击
		if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return fmt.Errorf("非法的文件路径: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			// 创建目录
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		// 创建文件
		if err := extractFile(file, filePath); err != nil {
			return err
		}
	}

	return nil
}

// extractFile 解压单个文件
func extractFile(file *zip.File, filePath string) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// 打开ZIP中的文件
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// 创建目标文件
	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 复制内容
	_, err = io.Copy(outFile, rc)
	return err
}

// WaitForProcessEnd 等待进程结束
func WaitForProcessEnd(timeoutMs int) {
	start := time.Now()
	for time.Since(start) < time.Duration(timeoutMs)*time.Millisecond {
		time.Sleep(100 * time.Millisecond)
	}
}

// GetProcessList 获取进程列表(用于检查进程是否存在)
func IsProcessRunning(processName string) bool {
	// 使用tasklist命令检查进程
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", processName))
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), processName)
}

// ForceKillProcess 强制终止进程
func ForceKillProcess(processName string) error {
	cmd := exec.Command("taskkill", "/IM", processName, "/F")
	return cmd.Run()
}

// GetProcessesByName 根据进程名获取PID列表
func GetProcessesByName(processName string) ([]string, error) {
	cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("name='%s'", processName), "get", "processid", "/format:csv")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var pids []string

	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) >= 2 && parts[1] != "" && parts[1] != "ProcessId" {
			pids = append(pids, strings.TrimSpace(parts[1]))
		}
	}

	return pids, nil
}
