package main

import (
	"herbarium/lib"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
)

type HerbariumApp struct {
	XDGName string
	App     *adw.Application
}

var appInstance *HerbariumApp

func GetHerbariumApp() *HerbariumApp {
	if appInstance == nil {
		xdgName := "ru.ximper.Herbarium"
		adwApp := adw.NewApplication(xdgName, gio.ApplicationFlagsNone)

		appInstance = &HerbariumApp{
			XDGName: xdgName,
			App:     adwApp,
		}
	}
	return appInstance
}

func (a *HerbariumApp) Activate() {
	mw := NewHerbariumWindow(a)
	mw.Window.Present()
}

func main() {
	lib.InitLocales()

	glib.LogSetDebugEnabled(false)

	app := GetHerbariumApp()

	app.App.ConnectActivate(func() {
		app.Activate()
	})

	os.Exit(app.App.Run(os.Args))
}
