package core

import (
	"fmt"
	"path/filepath"
	"strings"

	"club.xiaojiawei/hs-script-update/internal/config"
	"club.xiaojiawei/hs-script-update/internal/utils"
)

// Updater 更新器核心
type Updater struct {
	zipFilePath    string
	targetDir      string
	tempExtractDir string
}

// NewUpdater 创建更新器实例
func NewUpdater(zipFilePath, targetDir string) *Updater {
	return &Updater{
		zipFilePath:    zipFilePath,
		targetDir:      targetDir,
		tempExtractDir: filepath.Join(targetDir, "_temp_update"),
	}
}

// Update 执行更新
func (u *Updater) Update() error {
	fmt.Println("========================================")
	fmt.Println("开始更新程序")
	fmt.Println("========================================")
	fmt.Printf("更新包: %s\n", u.zipFilePath)
	fmt.Printf("目标目录: %s\n", u.targetDir)

	// 1. 检查更新包是否存在
	if !utils.Exists(u.zipFilePath) {
		return fmt.Errorf("更新包不存在: %s", u.zipFilePath)
	}

	// 2. 检查目标目录是否存在
	if !utils.Exists(u.targetDir) {
		return fmt.Errorf("目标目录不存在: %s", u.targetDir)
	}

	// 3. 清理临时目录
	if utils.Exists(u.tempExtractDir) {
		fmt.Println("清理旧的临时目录...")
		if err := utils.Delete(u.tempExtractDir); err != nil {
			return fmt.Errorf("清理临时目录失败: %w", err)
		}
	}

	// 4. 创建临时目录并解压
	if err := utils.CreateDirectory(u.tempExtractDir); err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}

	fmt.Println("\n解压更新包...")
	if err := utils.Unzip(u.zipFilePath, u.tempExtractDir); err != nil {
		u.cleanup()
		return fmt.Errorf("解压失败: %w", err)
	}

	// 5. 判断是 JVM 版还是 Native 版
	isJvmVersion := utils.DetectJVMVersion(u.targetDir)
	if isJvmVersion {
		fmt.Println("\n检测到版本类型: JVM")
	} else {
		fmt.Println("\n检测到版本类型: Native")
	}

	// 6. 执行更新
	if err := u.performUpdate(isJvmVersion); err != nil {
		u.cleanup()
		return err
	}

	// 7. 处理自我更新
	if err := utils.HandleSelfUpdate(u.tempExtractDir, u.targetDir); err != nil {
		fmt.Printf("警告: 更新器自更新失败: %v\n", err)
		fmt.Printf("请手动替换 %s\n", config.UpdaterName)
	}

	// 8. 清理临时目录
	fmt.Println("\n清理临时文件...")
	if err := utils.Delete(u.tempExtractDir); err != nil {
		fmt.Printf("警告: 清理临时目录失败: %v\n", err)
	}

	// 9. 删除更新包（如果在目标目录中）
	//if strings.HasPrefix(u.zipFilePath, u.targetDir) {
	//	fmt.Printf("删除更新包: %s\n", u.zipFilePath)
	//	if err := utils.Delete(u.zipFilePath); err != nil {
	//		fmt.Printf("警告: 删除更新包失败: %v\n", err)
	//	}
	//}

	fmt.Println("\n========================================")
	fmt.Println("更新完成！")
	fmt.Println("========================================")

	return nil
}

// performUpdate 执行更新操作
func (u *Updater) performUpdate(isJvmVersion bool) error {
	extractedDir := utils.FindExtractedDirectory(u.tempExtractDir)

	if isJvmVersion {
		return u.updateJVMVersion(extractedDir)
	}
	return u.updateNativeVersion(extractedDir)
}

// updateJVMVersion 更新 JVM 版本
func (u *Updater) updateJVMVersion(extractedDir string) error {
	fmt.Println("\n更新 JVM 版本...")
	fmt.Printf("保留目录: %s\n", strings.Join(config.JVMPreserveDirs, ", "))
	fmt.Printf("要更新的插件: %s\n", strings.Join(config.JVMUpdatePluginDirs, ", "))

	// 复制文件，排除需要保留的目录和插件
	if err := utils.CopyDirectory(
		extractedDir,
		u.targetDir,
		config.JVMPreserveDirs,
		config.JVMUpdatePluginDirs,
	); err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	fmt.Println("JVM 版本更新完成")
	return nil
}

// updateNativeVersion 更新 Native 版本
func (u *Updater) updateNativeVersion(extractedDir string) error {
	fmt.Println("\n更新 Native 版本...")
	fmt.Printf("保留目录: %s\n", strings.Join(config.NativePreserveDirs, ", "))

	// 复制文件，只排除 config 和 data
	if err := utils.CopyDirectory(
		extractedDir,
		u.targetDir,
		config.NativePreserveDirs,
		nil,
	); err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	fmt.Println("Native 版本更新完成")
	return nil
}

// cleanup 清理临时文件
func (u *Updater) cleanup() {
	if utils.Exists(u.tempExtractDir) {
		if err := utils.Delete(u.tempExtractDir); err != nil {
			fmt.Printf("清理临时目录失败: %v\n", err)
		}
	}
}
