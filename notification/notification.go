package notification

import (
	"asearch/config"
	"asearch/logger"

	"github.com/getlantern/systray"
	"github.com/pkg/errors"
	"github.com/skratchdot/open-golang/open"
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
	mConfig := systray.AddMenuItem("打开配置", "打开配置")
	mQuit := systray.AddMenuItem("退出", "退出工具")

	for {
		select {
		case <-mOpen.ClickedCh:
			err := open.Start("http://" + config.Conf.Addr)
			if err != nil {
				logger.Error(errors.Cause(err))
			}
		case <-mConfig.ClickedCh:
			err := open.Start(".")
			if err != nil {
				logger.Error(errors.Cause(err))
			}
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}
