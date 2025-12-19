package main

import (
	"flag"
	"fmt"
	"os"

	"club.xiaojiawei/hs-script-update/internal/core"
	"club.xiaojiawei/hs-script-update/internal/repository"
)

const version = "1.0.0"

func main() {
	helpFlag := flag.Bool("help", false, "显示帮助信息")

	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	latestCmd := flag.NewFlagSet("latest", flag.ExitOnError)

	checkDev := checkCmd.Bool("d", false, "检查开发版")
	checkNative := checkCmd.Bool("n", false, "Native 版本")
	checkInteractive := checkCmd.Bool("i", false, "交互模式（控制台显示）")
	checkRepo := checkCmd.String("r", "gitee", "仓库源 (gitee/github)")

	latestDev := latestCmd.Bool("d", false, "获取开发版")
	latestNative := latestCmd.Bool("n", false, "Native 版本")
	latestInteractive := latestCmd.Bool("i", false, "交互模式（控制台显示）")
	latestRepo := latestCmd.String("r", "gitee", "仓库源 (gitee/github)")

	// 如果没有参数，显示帮助
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	// 解析全局标志
	flag.Parse()
	if *helpFlag {
		showHelp()
		return
	}

	switch os.Args[1] {
	case "update":
		updateCmd.Parse(os.Args[2:])
		if updateCmd.NArg() < 2 {
			fmt.Println("错误: update 命令需要两个参数")
			fmt.Println("使用方法: hs-script-updater update <zipPath> <targetDir>")
			os.Exit(1)
		}
		zipPath := updateCmd.Arg(0)
		targetDir := updateCmd.Arg(1)
		handleUpdate(zipPath, targetDir)

	case "check":
		checkCmd.Parse(os.Args[2:])
		if checkCmd.NArg() < 1 {
			fmt.Println("错误: check 命令需要一个参数")
			fmt.Println("使用方法: hs-script-updater check <version> [-d] [-n] [-i]")
			os.Exit(1)
		}
		currentVersion := checkCmd.Arg(0)
		handleCheck(currentVersion, *checkDev, *checkNative, *checkInteractive, *checkRepo)

	case "latest":
		latestCmd.Parse(os.Args[2:])
		handleLatest(*latestDev, *latestNative, *latestInteractive, *latestRepo)

	case "--help", "-h", "help":
		showHelp()

	default:
		fmt.Printf("未知命令: %s\n", os.Args[1])
		fmt.Println("使用 --help 查看帮助信息")
		os.Exit(1)
	}
}

// handleUpdate 处理更新命令
func handleUpdate(zipPath, targetDir string) {
	updater := core.NewUpdater(zipPath, targetDir)
	if err := updater.Update(); err != nil {
		fmt.Printf("\n更新失败: %v\n", err)
		os.Exit(1)
	}
}

// createRepository 根据参数创建仓库实例
func createRepository(repoName string) repository.Repository {
	switch repoName {
	case "github":
		return repository.NewGitHubRepository()
	case "gitee":
		return repository.NewGiteeRepository()
	default:
		fmt.Printf("警告: 未知的仓库源 '%s'，使用默认仓库 gitee\n", repoName)
		return repository.NewGiteeRepository()
	}
}

// handleCheck 处理检查版本命令
func handleCheck(currentVersion string, dev, native, interactive bool, repoName string) {
	repo := createRepository(repoName)
	checker := core.NewVersionChecker(repo)

	result, err := checker.CheckVersion(currentVersion, dev, native, interactive)
	if err != nil {
		fmt.Printf("检查更新失败: %v\n", err)
		os.Exit(1)
	}

	// 非交互模式才输出 JSON
	if result != "" {
		fmt.Println(result)
	}
}

// handleLatest 处理获取最新版本命令
func handleLatest(dev, native, interactive bool, repoName string) {
	repo := createRepository(repoName)
	checker := core.NewVersionChecker(repo)

	result, err := checker.GetLatestVersion(dev, native, interactive)
	if err != nil {
		fmt.Printf("获取最新版本失败: %v\n", err)
		os.Exit(1)
	}

	// 非交互模式才输出 JSON
	if result != "" {
		fmt.Println(result)
	}
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Printf(`HS-Script 更新器 v%s

使用方法:
  hs-script-updater <command> [arguments]

命令:
  update <zipPath> <targetDir>              执行更新
  check <version> [-d] [-n] [-i] [-r repo]  检查版本更新（需要当前版本号）
  latest [-d] [-n] [-i] [-r repo]           获取最新版本信息

示例:
  # 执行更新
  hs-script-updater update "D:\hs-script_v4.13.0-GA.zip" "D:\hs-script"

  # 获取最新 JVM 版本（返回 JSON，默认 Gitee）
  hs-script-updater latest

  # 从 GitHub 获取最新版本
  hs-script-updater latest -r github

  # 获取最新 Native 版本
  hs-script-updater latest -n

  # 获取最新版本（交互模式）
  hs-script-updater latest -i

  # 获取最新开发版
  hs-script-updater latest -d

  # 检查 JVM 版更新（返回 JSON，默认 Gitee）
  hs-script-updater check "v4.13.0-GA"

  # 从 GitHub 检查更新
  hs-script-updater check "v4.13.0-GA" -r github

  # 检查 Native 版更新
  hs-script-updater check "v4.13.0-GA" -n

  # 检查更新（交互模式）
  hs-script-updater check "v4.13.0-GA" -i

选项:
  -d, --dev            检查/获取开发版
  -n, --native         Native 版本（默认为 JVM 版本）
  -i, --interactive    交互模式（控制台显示）
  -r, --repo           仓库源 (gitee/github，默认 gitee)
  -h, --help           显示帮助信息
`, version)
}
