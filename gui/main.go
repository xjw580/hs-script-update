package gui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"log"
	"syscall"
)

type MyWindow struct {
	mainWin     *walk.MainWindow
	progressBar *walk.ProgressBar
}

const (
	ProgramName = "hs-script更新程序"
	winWidth    = 600
	winHeight   = 250
)

var (
	mw         = new(MyWindow)
	programIco walk.Image
)

func init() {
	var err error
	programIco, err = walk.Resources.Image("favicon.ico")
	if err != nil {
		log.Println("favicon.ico读取失败")
	}
}

func ShowWindow() {
	//args := os.Args
	//if len(args) < 4 {
	//	return
	//}
	//var count = 0
	//for i, arg := range args {
	//	if strings.HasPrefix(arg, "--") {
	//		split := strings.Split(arg, "=")
	//		if i != 0 && len(split) > 1 {
	//			if split[1] == "--target" {
	//				count++
	//			} else if split[1] == "--source" {
	//				count++
	//			} else if split[1] == "--pause" {
	//				count++
	//			}
	//		}
	//	}
	//}
	//if count < 3 {
	//	return
	//}

	err := MainWindow{
		Title:    ProgramName,
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
			PushButton{
				Text: "Start",
				OnClicked: func() {
					go func() {
						//for i := 0; i <= 100; i++ {
						//	time.Sleep(50 * time.Millisecond)
						//}
						mw.progressBar.SetValue(mw.progressBar.Value() + 10)
						walk.MsgBox(mw.mainWin, "Done", "Progress completed", walk.MsgBoxIconInformation)
					}()
				},
			},
		},
	}.Create()
	if err != nil {
		log.Println(err)
		return
	}
	mw.mainWin.Run()
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
