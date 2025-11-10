package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/lxn/walk"
	"golang.org/x/sys/windows"
)

var (
	User32           = windows.NewLazySystemDLL("User32.dll")
	SendMessageW     = User32.NewProc("SendMessageW")
	GetSystemMetrics = User32.NewProc(`GetSystemMetrics`)
)

const (
	EM_LIMITTEXT = 0x00C5
)

func SetTextEditLimit(textEdit *walk.TextEdit, limit int) {
	hwnd := textEdit.Handle()
	SendMessageW.Call(uintptr(hwnd), EM_LIMITTEXT, uintptr(limit), 0)
}

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

// PidExists 指定PID是否存在
func PidExists(pid string) bool {
	// 构造命令：tasklist /FI "PID eq <pid>"
	cmd := exec.Command("cmd", "/C", "tasklist", "/FI", fmt.Sprintf("PID eq %s", pid))

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false
	}

	output := out.String()
	// 检查输出中是否包含 pid 对应的行（排除"无任务运行..."之类的提示）
	return strings.Contains(output, pid)
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

// IsDirEmpty 是否为空目录
func IsDirEmpty(dirPath string) bool {
	_, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true // 目录不存在
		}
		return false
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}
	return len(entries) == 0
}

// DeleteLibFiles 删除库文件
func DeleteLibFiles(folderPath string, excludeDirs []string, logFunc func(string)) error {
	return filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理文件
		if info.IsDir() {
			return nil
		}

		//// 检查是否是库文件
		//if !isLibraryFile(info.Name()) {
		//	return nil
		//}

		// 检查是否需要排除
		if shouldExclude(path, excludeDirs) {
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

func DeletePluginFiles(folderPath string, whitelistFileName []string, logFunc func(string)) error {
	return filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理文件
		if info.IsDir() {
			return nil
		}

		// 检查是否在白名单中(包含匹配)
		fileName := info.Name()
		for _, whiteFile := range whitelistFileName {
			if strings.Contains(fileName, whiteFile) {
				logFunc("跳过删除文件(白名单): " + path)
				return nil
			}
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

// getParentName 提取父目录的名称
func getParentName(path string) string {
	// 获取父目录路径
	parentPath := filepath.Dir(path)

	// 获取父目录的名字
	return filepath.Base(parentPath)
}

// shouldExclude 检查是否应该排除文件
func shouldExclude(path string, excludeDirs []string) bool {
	excludeMap := make(map[string]struct{}, len(excludeDirs))
	for _, d := range excludeDirs {
		excludeMap[d] = struct{}{}
	}

	dir := filepath.Dir(path)
	for {
		base := filepath.Base(dir)
		if _, found := excludeMap[base]; found {
			return true
		}

		parent := filepath.Dir(dir)
		// 到达根目录或无变化时终止
		if parent == dir {
			break
		}
		dir = parent
	}
	return false
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

// ProcessKiller 进程终止器
type ProcessKiller struct {
	logger *log.Logger
}

// NewProcessKiller 创建新的进程终止器
func NewProcessKiller() *ProcessKiller {
	return &ProcessKiller{
		logger: log.New(os.Stdout, "[ProcessKiller] ", log.LstdFlags),
	}
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID  int    `json:"pid"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// KillProcessByName 根据进程名杀死进程
func (pk *ProcessKiller) KillProcessByName(processName string) error {
	switch runtime.GOOS {
	case "windows":
		return pk.killProcessWindows(processName)
	case "linux", "darwin":
		return pk.killProcessUnix(processName)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// killProcessWindows Windows 系统杀死进程
func (pk *ProcessKiller) killProcessWindows(processName string) error {
	// 方法1: 使用 taskkill 命令（推荐）
	err := pk.killByTaskkill(processName)
	if err == nil {
		pk.logger.Printf("使用 taskkill 成功杀死进程: %s", processName)
		return nil
	}

	pk.logger.Printf("taskkill 失败，尝试备用方法: %v", err)

	// 方法2: 获取进程列表后逐个杀死
	return pk.killByProcessList(processName)
}

// killByTaskkill 使用 taskkill 命令
func (pk *ProcessKiller) killByTaskkill(processName string) error {
	// 确保进程名包含 .exe 后缀
	if !strings.HasSuffix(strings.ToLower(processName), ".exe") {
		processName += ".exe"
	}

	// 使用 taskkill /f /im 强制杀死进程
	cmd := exec.Command("taskkill", "/f", "/im", processName)
	output, err := cmd.CombinedOutput()

	pk.logger.Printf("taskkill 输出: %s", string(output))

	if err != nil {
		return fmt.Errorf("taskkill 执行失败: %v, 输出: %s", err, string(output))
	}

	// 检查输出是否表示成功
	outputStr := string(output)
	if strings.Contains(outputStr, "成功") || strings.Contains(outputStr, "SUCCESS") {
		return nil
	}

	return fmt.Errorf("taskkill 未找到进程或执行失败: %s", outputStr)
}

// killByProcessList 通过进程列表查找并杀死
func (pk *ProcessKiller) killByProcessList(processName string) error {
	processes, err := pk.getWindowsProcesses(processName)
	if err != nil {
		return fmt.Errorf("获取进程列表失败: %v", err)
	}

	if len(processes) == 0 {
		return fmt.Errorf("未找到进程: %s", processName)
	}

	var lastError error
	killedCount := 0

	for _, proc := range processes {
		err := pk.killProcessByPID(proc.PID)
		if err != nil {
			pk.logger.Printf("杀死进程 %d (%s) 失败: %v", proc.PID, proc.Name, err)
			lastError = err
		} else {
			pk.logger.Printf("成功杀死进程: PID=%d, Name=%s", proc.PID, proc.Name)
			killedCount++
		}
	}

	if killedCount == 0 {
		return fmt.Errorf("所有进程杀死都失败，最后错误: %v", lastError)
	}

	pk.logger.Printf("成功杀死 %d/%d 个进程", killedCount, len(processes))
	return nil
}

// getWindowsProcesses 获取 Windows 进程列表
func (pk *ProcessKiller) getWindowsProcesses(processName string) ([]ProcessInfo, error) {
	// 使用 wmic 获取进程信息
	cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("name='%s'", processName),
		"get", "processid,name,executablepath", "/format:csv")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行 wmic 命令失败: %v", err)
	}

	return pk.parseWmicOutput(string(output))
}

// parseWmicOutput 解析 wmic 输出
func (pk *ProcessKiller) parseWmicOutput(output string) ([]ProcessInfo, error) {
	lines := strings.Split(output, "\n")
	var processes []ProcessInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Node") || strings.HasPrefix(line, "ExecutablePath") {
			continue
		}

		// CSV 格式: Node,ExecutablePath,Name,ProcessId
		parts := strings.Split(line, ",")
		if len(parts) >= 4 {
			pidStr := strings.TrimSpace(parts[3])
			if pid, err := strconv.Atoi(pidStr); err == nil {
				processes = append(processes, ProcessInfo{
					PID:  pid,
					Name: strings.TrimSpace(parts[2]),
					Path: strings.TrimSpace(parts[1]),
				})
			}
		}
	}

	return processes, nil
}

// killProcessUnix Unix/Linux 系统杀死进程
func (pk *ProcessKiller) killProcessUnix(processName string) error {
	// 使用 pkill 命令
	cmd := exec.Command("pkill", "-f", processName)
	output, err := cmd.CombinedOutput()

	pk.logger.Printf("pkill 输出: %s", string(output))

	if err != nil {
		// pkill 没找到进程时也会返回错误，需要检查具体错误
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return fmt.Errorf("未找到进程: %s", processName)
			}
		}
		return fmt.Errorf("pkill 执行失败: %v", err)
	}

	pk.logger.Printf("成功杀死进程: %s", processName)
	return nil
}

// killProcessByPID 根据 PID 杀死进程
func (pk *ProcessKiller) killProcessByPID(pid int) error {
	cmd := exec.Command("taskkill", "/f", "/pid", strconv.Itoa(pid))
	_, err := cmd.Output()
	return err
}

// ForceKillProcessByName 强制杀死进程（不可恢复）
func (pk *ProcessKiller) ForceKillProcessByName(processName string) error {
	switch runtime.GOOS {
	case "windows":
		return pk.forceKillWindows(processName)
	case "linux", "darwin":
		return pk.forceKillUnix(processName)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// forceKillWindows Windows 强制杀死
func (pk *ProcessKiller) forceKillWindows(processName string) error {
	if !strings.HasSuffix(strings.ToLower(processName), ".exe") {
		processName += ".exe"
	}

	// 使用 /f 参数强制终止
	cmd := exec.Command("taskkill", "/f", "/im", processName, "/t")
	output, err := cmd.CombinedOutput()

	pk.logger.Printf("强制杀死输出: %s", string(output))
	return err
}

// forceKillUnix Unix 强制杀死
func (pk *ProcessKiller) forceKillUnix(processName string) error {
	// 使用 SIGKILL 信号强制终止
	cmd := exec.Command("pkill", "-9", "-f", processName)
	output, err := cmd.CombinedOutput()

	pk.logger.Printf("强制杀死输出: %s", string(output))
	return err
}

// GetProcessesByName 根据进程名获取进程列表
func (pk *ProcessKiller) GetProcessesByName(processName string) ([]ProcessInfo, error) {
	switch runtime.GOOS {
	case "windows":
		return pk.getWindowsProcesses(processName)
	case "linux", "darwin":
		return pk.getUnixProcesses(processName)
	default:
		return nil, fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// getUnixProcesses 获取 Unix 进程列表
func (pk *ProcessKiller) getUnixProcesses(processName string) ([]ProcessInfo, error) {
	cmd := exec.Command("pgrep", "-f", processName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pgrep 执行失败: %v", err)
	}

	var processes []ProcessInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if pid, err := strconv.Atoi(strings.TrimSpace(line)); err == nil {
			processes = append(processes, ProcessInfo{
				PID:  pid,
				Name: processName,
				Path: "", // pgrep 不返回路径
			})
		}
	}

	return processes, nil
}

// IsProcessRunning 检查进程是否正在运行
func (pk *ProcessKiller) IsProcessRunning(processName string) bool {
	processes, err := pk.GetProcessesByName(processName)
	if err != nil {
		pk.logger.Printf("检查进程状态失败: %v", err)
		return false
	}
	return len(processes) > 0
}

// KillAllProcessesByName 杀死所有匹配的进程
func (pk *ProcessKiller) KillAllProcessesByName(processName string) (int, error) {
	processes, err := pk.GetProcessesByName(processName)
	if err != nil {
		return 0, err
	}

	if len(processes) == 0 {
		return 0, fmt.Errorf("未找到进程: %s", processName)
	}

	killedCount := 0
	for _, proc := range processes {
		if err := pk.killProcessByPID(proc.PID); err == nil {
			killedCount++
			pk.logger.Printf("杀死进程: PID=%d", proc.PID)
		} else {
			pk.logger.Printf("杀死进程 %d 失败: %v", proc.PID, err)
		}
	}

	return killedCount, nil
}
