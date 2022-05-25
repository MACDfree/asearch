package notification

import (
	"asearch/config"
	"asearch/logger"
	"asearch/util"

	"github.com/getlantern/systray"
)

func Run() {
	onExit := func() {
		logger.Info("退出...")
	}
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTemplateIcon(icon, icon)
	systray.SetTitle("本地全文检索工具")
	systray.SetTooltip("本地全文检索工具")
	mOpen := systray.AddMenuItem("打开网页", "打开网页")
	mQuit := systray.AddMenuItem("退出", "退出工具")

	for {
		select {
		case <-mOpen.ClickedCh:
			util.OpenLocal("http://" + config.Conf.Addr)
		case <-mQuit.ClickedCh:
			systray.Quit()
			logger.Info("退出...")
			return
		}
	}
}
