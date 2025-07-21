package gui

import (
	"fmt"
	"hs-script-update/config"
	"log"
	"syscall"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// GUI 图形界面结构体
type GUI struct {
	mainWindow  *walk.MainWindow
	progressBar *walk.ProgressBar
	logTextEdit *walk.TextEdit
	logCallback func(string)
	statusChan  chan error
}

// NewGUI 创建新的GUI实例
func NewGUI(logCallback func(string)) *GUI {
	return &GUI{
		logCallback: logCallback,
		statusChan:  make(chan error, 1),
	}
}

// Show 显示GUI并执行更新任务
func (g *GUI) Show(updateFunc func() error) error {
	go g.createWindow(updateFunc)
	return <-g.statusChan
}

// createWindow 创建窗口
func (g *GUI) createWindow(updateFunc func() error) {
	err := MainWindow{
		Title:    config.ProgramName,
		AssignTo: &g.mainWindow,
		Bounds: Rectangle{
			X:      int(g.getDisplayWidth()-config.WindowWidth) >> 1,
			Y:      int(g.getDisplayHeight()-config.WindowHeight-40) >> 1,
			Width:  config.WindowWidth,
			Height: config.WindowHeight,
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
				AssignTo: &g.progressBar,
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
				AssignTo: &g.logTextEdit,
				VScroll:  true,
				ReadOnly: true,
			},
		},
	}.Create()

	if err != nil {
		log.Printf("创建窗口失败: %v", err)
		g.statusChan <- err
		return
	}

	// 启动更新任务
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("更新任务异常: %v", r)
				g.statusChan <- err
			}
		}()

		err := updateFunc()
		g.statusChan <- err
	}()

	// 运行GUI消息循环
	g.mainWindow.Run()
}

// SetProgress 设置进度条值
func (g *GUI) SetProgress(value int) {
	if g.progressBar != nil {
		g.progressBar.SetValue(value)
	}
}

// Log 记录日志
func (g *GUI) Log(message string) {
	if g.logTextEdit != nil {
		g.logTextEdit.AppendText(message + "\r\n")
	}
	if g.logCallback != nil {
		g.logCallback(message)
	}
}

// Logf 格式化记录日志
func (g *GUI) Logf(format string, args ...interface{}) {
	g.Log(fmt.Sprintf(format, args...))
}

// IsVisible 检查窗口是否可见
func (g *GUI) IsVisible() bool {
	return g.mainWindow != nil && g.mainWindow.Visible()
}

// getDisplayWidth 获取显示器宽度
func (g *GUI) getDisplayWidth() uintptr {
	w, _, _ := syscall.NewLazyDLL(`User32.dll`).NewProc(`GetSystemMetrics`).Call(uintptr(0))
	return w
}

// getDisplayHeight 获取显示器高度
func (g *GUI) getDisplayHeight() uintptr {
	h, _, _ := syscall.NewLazyDLL(`User32.dll`).NewProc(`GetSystemMetrics`).Call(uintptr(1))
	return h
}
