package gui

import (
	"Hearthstone-Script-update/util"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type MyWindow struct {
	mainWin     *walk.MainWindow
	progressBar *walk.ProgressBar
	logTextEdit *walk.TextEdit
}

const (
	sourceProgramName = "hs-script"
	programName       = sourceProgramName + "更新程序"
	winWidth          = 600
	winHeight         = 400
	defaultVersionDir = "new_version_temp"
)

var (
	mw             = new(MyWindow)
	target         string
	source         string
	pause          string
	pid            string
	statusChan     chan int
	versionZipFile string
)

func argsCheck() bool {
	args := os.Args
	for i, arg := range args {
		if strings.HasPrefix(arg, "--") {
			split := strings.Split(arg, "=")
			if i != 0 && len(split) > 1 {
				if split[0] == "--target" {
					target = strings.Trim(split[1], "\"")
					target = strings.Trim(split[1], "'")
				} else if split[0] == "--source" {
					source = strings.Trim(split[1], "\"")
					source = strings.Trim(split[1], "'")
				} else if split[0] == "--pause" {
					pause = strings.Trim(split[1], "\"")
					pause = strings.Trim(split[1], "'")
				} else if split[0] == "--pid" {
					pid = strings.Trim(split[1], "\"")
					pid = strings.Trim(split[1], "'")
				}
			}
		}
	}
	if target == "" {
		executable, _ := os.Executable()
		target = filepath.Dir(executable)
	}
	if source == "" {
		executable, _ := os.Executable()
		currentDir := filepath.Dir(executable)
		versionDir := filepath.Join(currentDir, defaultVersionDir)
		s, err := os.Stat(versionDir)
		if err == nil && s.IsDir() && getNewestSubdirectory(versionDir) != "" {
			source = getNewestSubdirectory(versionDir)
		} else {
			err = filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if path == currentDir {
					return nil
				}

				if info.IsDir() {
					return filepath.SkipDir // 跳过子目录
				}
				if strings.HasSuffix(info.Name(), ".zip") {
					versionZipFile = info.Name()
				}
				return nil
			})
			if versionZipFile == "" {
				return false
			}
			zipFilePath := filepath.Join(currentDir, versionZipFile)
			fileName := filepath.Base(zipFilePath)

			// 去掉文件后缀
			nameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
			source = filepath.Join(versionDir, nameWithoutExt)
			err := util.Unzip(zipFilePath, source)
			if err != nil {
				source = ""
				return false
			}
			return true
		}
	}
	return true
}

func getNewestSubdirectory(dirPath string) string {
	var newestDir string
	var latestTime int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 检查是否是目录且不是根目录
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

func ShowWindow() {
	statusChan = make(chan int, 3)
	go func() {
		err := MainWindow{
			Title:    programName,
			AssignTo: &mw.mainWin,
			Bounds: Rectangle{
				X:      int(getDisplayWidth()-winWidth) >> 1,
				Y:      int(getDisplayHeight()-winHeight-40) >> 1,
				Width:  winWidth,
				Height: winHeight,
			},
			Font: Font{
				PointSize: 11,
			},
			Background: SolidColorBrush{
				Color: walk.RGB(224, 240, 253),
			},
			Layout: VBox{},
			Children: []Widget{
				ProgressBar{
					AssignTo: &mw.progressBar,
					MinValue: 0,
					MaxValue: 100,
					MaxSize: Size{
						Height: 15,
					},
				},
				TextLabel{
					Text: "详细信息",
				},
				TextEdit{
					AssignTo: &mw.logTextEdit,
					VScroll:  true,
					ReadOnly: true,
				},
			},
		}.Create()
		if err != nil {
			log.Println(err)
			statusChan <- 0
			return
		}
		go func() {
			if !argsCheck() {
				statusChan <- 0
				return
			}
			execUpdate()
		}()
		mw.mainWin.Run()
	}()
	<-statusChan
}

func execUpdate() {
	if pid != "" {
		appendLog("开始关闭" + sourceProgramName)
		_ = util.KillProcess(pid)
		time.Sleep(time.Millisecond * 1000)
		appendLog("已关闭" + sourceProgramName)
		mw.progressBar.SetValue(mw.progressBar.Value() + 5)
	}
	appendLog("==========》开始删除无用文件《==========")
	delLibFile(mw.progressBar, 50)
	appendLog("==========》开始复制更新文件《==========")
	_ = util.CopyDirectory(source, target, mw.progressBar, func(log string) {
		appendLog(log)
	}, 90)
	appendLog("==========》复制更新文件完毕《==========")
	//if strings.Contains(source, "new_version_temp") {
	//	appendLog("删除版本文件 " + source)
	//	err := os.RemoveAll(source)
	//	mw.progressBar.SetValue(mw.progressBar.Value() + 5)
	//	if err != nil {
	//		log.Println("Error:", err)
	//	}
	//}
	appendLog("==========》开始启动" + sourceProgramName + "《==========")
	sourceProgramPath := target + "/" + sourceProgramName + ".exe"
	exists := util.FileExists(sourceProgramPath)
	if exists {
		_ = exec.Command(sourceProgramPath, "--pause="+pause).Run()
		appendLog("==========》" + sourceProgramName + "启动完毕《==========")
	} else {
		appendLog("启动失败，未找到" + sourceProgramPath)
	}
	mw.progressBar.SetValue(100)
	closeTime := 10
	for i := range closeTime {
		if !mw.mainWin.Visible() {
			break
		}
		appendLog(strconv.Itoa(closeTime-i) + "秒后关闭本程序")
		time.Sleep(time.Second * 1)
	}
	statusChan <- 0
}

func delLibFile(progress *walk.ProgressBar, maxProgressValue int) {
	if util.FileExists(target + "/" + sourceProgramName + ".exe") {
		appendLog("==========》开始删除旧依赖文件《==========")
		err := util.DeleteLibFiles(target, filepath.Base(source), progress, func(log string) {
			appendLog(log)
		}, maxProgressValue)
		time.Sleep(time.Millisecond * 1000)
		mw.progressBar.SetValue(mw.progressBar.Value() + 20)
		if err != nil {
			log.Println(err)
		}
	}
}

/*
*
获取显示器宽度
*/
func getDisplayWidth() uintptr {
	w, _, _ := syscall.NewLazyDLL(`User32.dll`).NewProc(`GetSystemMetrics`).Call(uintptr(0))
	return w
}

/*
*
获取显示器高度
*/
func getDisplayHeight() uintptr {
	h, _, _ := syscall.NewLazyDLL(`User32.dll`).NewProc(`GetSystemMetrics`).Call(uintptr(1))
	return h
}

func appendLog(log string) {
	mw.logTextEdit.AppendText(log + "\r\n")
}
