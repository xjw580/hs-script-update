package config

import (
	"fmt"
	"hs-script-update/utils"
	"os"
	"path/filepath"
	"strings"
)

const (
	SourceProgramName   = "hs-script"
	UpdaterName         = "update.exe"
	ProgramName         = SourceProgramName + "更新程序"
	DefaultVersionDir   = "new_version_temp"
	TempUpdaterSuffix   = ".temp"
	WindowWidth         = 600
	WindowHeight        = 400
	DefaultCloseDelay   = 10
	SelfUpdateParam     = "-self-update"
	TargetParam         = "--target"
	SourceParam         = "--source"
	PauseParam          = "--pause"
	PidParam            = "--pid"
	VersionZipFileParam = "--version-file"
)

// Config 配置结构体
type Config struct {
	TargetDir       string // 目标目录
	SourceDir       string // 源目录
	PauseFlag       string // 暂停标志
	ProcessID       string // 进程ID
	VersionZipFile  string // 版本压缩包文件
	TempUpdaterPath string // 临时更新器路径
}

// ParseConfig 解析命令行参数和环境配置
func ParseConfig() (*Config, error) {
	config := &Config{}

	// 解析命令行参数
	if err := config.parseArgs(); err != nil {
		return nil, err
	}

	// 自动检测配置
	if err := config.autoDetect(); err != nil {
		return nil, err
	}

	if utils.IsDirEmpty(config.SourceDir) {
		return nil, fmt.Errorf("SourceDir [%s] 目录为空或不存在", config.SourceDir)
	}

	return config, nil
}

// parseArgs 解析命令行参数
func (c *Config) parseArgs() error {
	args := os.Args
	for i, arg := range args {
		if strings.HasPrefix(arg, "--") && i != 0 {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := strings.Trim(parts[1], "\"'")

				switch key {
				case TargetParam:
					c.TargetDir = value
				case SourceParam:
					c.SourceDir = value
				case PauseParam:
					c.PauseFlag = value
				case PidParam:
					c.ProcessID = value
				case VersionZipFileParam:
					c.VersionZipFile = value
				}
			}
		}
	}
	return nil
}

// autoDetect 自动检测配置
func (c *Config) autoDetect() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	currentDir := filepath.Dir(executable)

	// 设置默认目标目录
	if c.TargetDir == "" {
		c.TargetDir = currentDir
	}

	// 自动检测源目录
	if c.SourceDir == "" {
		if err := c.detectSourceDir(currentDir); err != nil {
			return err
		}
	}

	// 检查是否需要自我更新
	err = c.checkSelfUpdate()
	if err != nil {
		return err
	}

	return nil
}

// checkSelfUpdate 检查是否需要自我更新
func (c *Config) checkSelfUpdate() error {
	var selfUpdate = false
	for _, arg := range os.Args {
		if arg == SelfUpdateParam {
			selfUpdate = true
		}
	}
	if !selfUpdate {
		// 检查源目录中是否有新的更新器
		newUpdaterPath := filepath.Join(c.SourceDir, UpdaterName)
		if utils.FileExists(newUpdaterPath) {
			c.TempUpdaterPath = filepath.Join(os.TempDir(), SourceProgramName, UpdaterName)
			// 创建临时更新器路径
			if err := utils.CopyFile(newUpdaterPath, c.TempUpdaterPath); err != nil {
				return fmt.Errorf("复制更新器失败: %v", err)
			}
		} else {
			return fmt.Errorf("找不到更新器%s", newUpdaterPath)
		}
	}

	return nil
}

// isUpdaterNewer 检查新更新器是否更新
func (c *Config) isUpdaterNewer(newPath, currentPath string) bool {
	newInfo, err1 := os.Stat(newPath)
	currentInfo, err2 := os.Stat(currentPath)

	if err1 != nil || err2 != nil {
		return false
	}

	return newInfo.ModTime().After(currentInfo.ModTime())
}

// detectSourceDir 检测源目录
func (c *Config) detectSourceDir(currentDir string) error {
	versionDir := filepath.Join(currentDir, DefaultVersionDir)

	// 检查版本目录中的最新子目录
	//if newestDir := getNewestSubdirectory(versionDir); newestDir != "" {
	//	c.SourceDir = newestDir
	//	return nil
	//}

	c.SourceDir = versionDir

	var zipFileName string
	if c.VersionZipFile == "" {
		// 查找zip文件
		zipFile, err := c.findZipFile(currentDir)
		if err != nil {
			return err
		}
		zipFileName = zipFile
		c.VersionZipFile = zipFileName
	}

	if zipFileName != "" {
		// 解压zip文件
		return c.extractZipFile(currentDir, versionDir, zipFileName)
	}

	return fmt.Errorf("未找到有效的源目录或zip文件")
}

// findZipFile 查找zip文件
func (c *Config) findZipFile(dir string) (string, error) {
	var zipFile string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == dir {
			return nil
		}
		if info.IsDir() {
			return filepath.SkipDir
		}
		if strings.HasSuffix(info.Name(), ".zip") {
			zipFile = info.Name()
		}
		return nil
	})
	return zipFile, err
}

// extractZipFile 解压zip文件
func (c *Config) extractZipFile(currentDir, versionDir, zipFile string) error {
	zipFilePath := filepath.Join(currentDir, zipFile)
	//fileName := filepath.Base(zipFilePath)
	//nameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	//c.SourceDir = filepath.Join(versionDir, nameWithoutExt)
	c.SourceDir = versionDir

	return utils.UnzipFile(zipFilePath, c.SourceDir)
}

// getNewestSubdirectory 获取最新的子目录
func getNewestSubdirectory(dirPath string) string {
	var newestDir string
	var latestTime int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != dirPath {
			if info.ModTime().Unix() > latestTime {
				latestTime = info.ModTime().Unix()
				newestDir = path
			}
		}
		return nil
	})

	if err != nil {
		return ""
	}
	return newestDir
}
