package gui

import (
	"Hearthstone-Script-update/util"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"log"
	"os"
	"os/exec"
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
	winHeight         = 300
)

var (
	mw         = new(MyWindow)
	programIco walk.Image
	target     string
	source     string
	pause      string
	pid        string
	statusChan chan int
)

func init() {
	var err error
	programIco, err = walk.Resources.Image("favicon.ico")
	if err != nil {
		log.Println("favicon.ico读取失败")
	}
}

func ShowWindow() {
	args := os.Args
	if len(args) < 4 {
		return
	}
	for i, arg := range args {
		if strings.HasPrefix(arg, "--") {
			split := strings.Split(arg, "=")
			if i != 0 && len(split) > 1 {
				if split[0] == "--target" {
					target = split[1]
				} else if split[0] == "--source" {
					source = split[1]
				} else if split[0] == "--pause" {
					pause = split[1]
				} else if split[0] == "--pid" {
					pid = split[1]
				}
			}
		}
	}
	if target == "" || source == "" {
		log.Println("ERROR: target或source参数为空")
		return
	}
	statusChan = make(chan int, 3)
	go func() {
		err := MainWindow{
			Title:    programName,
			Icon:     programIco,
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
		pid = "12345"
		if err != nil {
			log.Println(err)
			return
		}
		go execUpdate()
		mw.mainWin.Run()
	}()
	<-statusChan
}

func execUpdate() {
	if pid != "" {
		appendLog("开始关闭" + sourceProgramName)
		_ = util.KillProcess(pid)
		time.Sleep(time.Millisecond * 500)
		appendLog("已关闭" + sourceProgramName)
		mw.progressBar.SetValue(mw.progressBar.Value() + 5)
	}
	appendLog("==========》开始复制更新文件《==========")
	_ = util.CopyDirectory(source, target, mw.progressBar, func(log string) {
		appendLog(log)
	})
	appendLog("==========》复制更新文件完毕《==========")
	appendLog("==========》开始启动" + sourceProgramName + "《==========")
	sourceProgramPath := source + "/" + sourceProgramName + ".exe"
	exists := util.FileExists(sourceProgramName)
	if exists {
		_ = exec.Command(sourceProgramPath, "--pause="+pause).Run()
		appendLog("==========》" + sourceProgramName + "启动完毕《==========")
	} else {
		appendLog("启动失败，未找到" + sourceProgramName)
	}
	mw.progressBar.SetValue(100)
	statusChan <- 0
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
