package main

import (
	"fmt"
	"hs-script-update/config"
	"hs-script-update/updater"
	"hs-script-update/utils"
	"log"
	"os"
)

func main() {
	// 初始化日志
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 检查是否是自我更新模式
	if len(os.Args) > 1 && os.Args[1] == "--self-update" {
		if err := handleSelfUpdate(); err != nil {
			log.Fatalf("自我更新失败: %v", err)
		}
		return
	}

	// 解析命令行参数
	parseConfig, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("参数解析失败: %v", err)
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
func handleSelfUpdate() error {
	if len(os.Args) < 4 {
		return fmt.Errorf("自我更新参数不足")
	}

	tempUpdater := os.Args[2]
	targetUpdater := os.Args[3]

	log.Printf("开始自我更新: %s -> %s", tempUpdater, targetUpdater)

	// 等待原更新器退出
	utils.WaitForProcessEnd(5000) // 等待5秒

	// 复制新的更新器
	if err := utils.CopyFile(tempUpdater, targetUpdater); err != nil {
		return fmt.Errorf("复制更新器失败: %v", err)
	}

	// 删除临时文件
	os.Remove(tempUpdater)

	log.Println("自我更新完成")
	return nil
}
