package updater

import (
	"fmt"
	"hs-script-update/config"
	"hs-script-update/gui"
	"hs-script-update/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Updater 更新器结构体
type Updater struct {
	config *config.Config
	gui    *gui.GUI
	logger Logger
}

// Logger 日志接口
type Logger interface {
	Log(message string)
	Logf(format string, args ...interface{})
}

// NewUpdater 创建新的更新器实例
func NewUpdater(config *config.Config) *Updater {
	updater := &Updater{
		config: config,
	}

	// 创建GUI
	mainGUI := gui.NewGUI(updater.onLog)
	updater.gui = mainGUI
	updater.logger = mainGUI

	return updater
}

// Run 运行更新器
func (u *Updater) Run() error {
	// 启动GUI
	return u.gui.Show(func() error {
		return u.executeUpdate()
	})
}

// executeUpdate 执行更新流程
func (u *Updater) executeUpdate() error {
	u.logger.Logf("PID: %s", u.config.ProcessID)
	u.logger.Logf("SourceDir: %s", u.config.SourceDir)
	u.logger.Logf("TargetDir: %s", u.config.TargetDir)
	u.logger.Logf("PauseFlag: %s", u.config.PauseFlag)
	u.logger.Logf("VersionZipFile: %s", u.config.VersionZipFile)

	steps := []struct {
		name string
		fn   func() error
	}{
		{"关闭目标程序", u.stopTargetProgram},
		{"删除旧文件", u.cleanOldFiles},
		{"复制更新文件", u.copyUpdateFiles},
		{"启动目标程序", u.startTargetProgram},
		{"清理临时文件", u.cleanup},
	}

	totalSteps := len(steps)
	for i, step := range steps {
		u.logger.Logf("==========》开始%s《==========", step.name)

		if err := step.fn(); err != nil {
			u.logger.Logf("%s失败: %v", step.name, err)
			return err
		}

		u.logger.Logf("==========》%s完毕《==========", step.name)
		progress := int((float64(i+1) / float64(totalSteps)) * 100) // 留10%给倒计时
		u.gui.SetProgress(progress)
	}

	// 倒计时关闭
	u.countdown()

	return nil
}

// stopTargetProgram 停止目标程序
func (u *Updater) stopTargetProgram() error {
	killer := utils.NewProcessKiller()
	var platformName = "Battle.net.exe"
	var gameName = "Hearthstone.exe"
	if killer.IsProcessRunning(gameName) {
		err := killer.ForceKillProcessByName(gameName)
		if err != nil {
			u.logger.Logf("关闭%s失败: %v", gameName, err)
		}
	}
	if killer.IsProcessRunning(platformName) {
		err := killer.ForceKillProcessByName(platformName)
		if err != nil {
			u.logger.Logf("关闭%s失败: %v", platformName, err)
		}
	}

	if !utils.PidExists(u.config.ProcessID) {
		u.logger.Logf("指定进程ID[%s]不存在，跳过关闭程序", u.config.ProcessID)
		return nil
	}

	u.logger.Logf("开始关闭%s (PID: %s)", config.SourceProgramName, u.config.ProcessID)

	if err := utils.KillProcess(u.config.ProcessID); err != nil {
		u.logger.Logf("关闭程序失败: %v", err)
		return err
	}

	time.Sleep(time.Second)
	u.logger.Logf("已关闭%s", config.SourceProgramName)
	return nil
}

// cleanOldFiles 清理旧文件
func (u *Updater) cleanOldFiles() error {
	targetExe := filepath.Join(u.config.TargetDir, config.SourceProgramName+".exe")
	if !utils.FileExists(targetExe) {
		u.logger.Log("目标程序不存在，跳过删除旧依赖文件")
		return nil
	}

	excludeDir := filepath.Base(u.config.SourceDir)
	return utils.DeleteLibFiles(
		u.config.TargetDir,
		[]string{excludeDir, "plugin", "config", "data", "log"},
		func(message string) { u.logger.Log(message) },
	)
}

// copyUpdateFiles 复制更新文件
func (u *Updater) copyUpdateFiles() error {
	return utils.CopyDirectory(
		u.config.SourceDir,
		u.config.TargetDir,
		func(message string) { u.logger.Log(message) },
	)
}

// startTargetProgram 启动目标程序
func (u *Updater) startTargetProgram() error {
	time.Sleep(time.Second)
	targetExe := filepath.Join(u.config.TargetDir, config.SourceProgramName+".exe")

	if !utils.FileExists(targetExe) {
		u.logger.Logf("启动失败，未找到%s", targetExe)
		return nil
	}

	var args []string
	if u.config.PauseFlag != "" {
		args = append(args, "--pause="+u.config.PauseFlag)
	}

	cmd := exec.Command("cmd", "/C", "start", "", targetExe)
	cmd.Args = append(cmd.Args, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	err := cmd.Start()
	if err != nil {
		u.logger.Logf("启动程序失败: %v", err)
		return nil
	}

	u.logger.Logf("==========》%s启动完毕《==========", config.SourceProgramName)
	return nil
}

// cleanup 清理临时文件
func (u *Updater) cleanup() error {
	if strings.Contains(u.config.SourceDir, config.DefaultVersionDir) {
		u.logger.Logf("删除临时文件: %s", u.config.SourceDir)
		if err := os.RemoveAll(u.config.SourceDir); err != nil {
			u.logger.Logf("删除临时文件失败: %v", err)
			return err
		}
	}
	return nil
}

// countdown 倒计时关闭
func (u *Updater) countdown() {
	u.logger.Logf("%d秒后关闭本程序", config.DefaultCloseDelay)
	for i := config.DefaultCloseDelay; i > 0; i-- {
		if !u.gui.IsVisible() {
			break
		}
		time.Sleep(time.Second)
	}
}

// onLog 日志回调
func (u *Updater) onLog(message string) {
	fmt.Println(message)
}
