package main

import (
	"fmt"
	"hs-script-update/config"
	"hs-script-update/updater"
	"log"
	"os"
	"os/exec"
	"time"
)

func main() {
	// 初始化日志
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 解析命令行参数
	parseConfig, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("参数解析失败: %v", err)
	}

	if parseConfig.TempUpdaterPath != "" {
		if err := handleSelfUpdate(parseConfig); err != nil {
			log.Fatalf("自我更新失败: %v", err)
		}
		return
	} else {
		time.Sleep(time.Second)
	}

	// 创建更新器实例
	newUpdater := updater.NewUpdater(parseConfig)

	// 运行更新器
	if err := newUpdater.Run(); err != nil {
		log.Printf("更新失败: %v", err)
		fmt.Println("按任意键退出...")
		fmt.Scanln()
		os.Exit(1)
	}

	log.Println("更新完成")
}

// handleSelfUpdate 处理自我更新
func handleSelfUpdate(parseConfig *config.Config) error {

	var args []string
	args = append(args, config.TargetParam+"='"+parseConfig.TargetDir+"'")
	args = append(args, config.SourceParam+"='"+parseConfig.SourceDir+"'")
	args = append(args, config.PauseParam+"='"+parseConfig.PauseFlag+"'")
	args = append(args, config.PidParam+"='"+parseConfig.ProcessID+"'")
	args = append(args, config.SelfUpdateParam)
	cmd := exec.Command("cmd", "/C", "start", "", parseConfig.TempUpdaterPath)
	cmd.Args = append(cmd.Args, args...)

	err := cmd.Start()
	if err != nil {
		return err
	}
	return nil
}
