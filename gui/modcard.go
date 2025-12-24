package main

import (
	"bytes"
	"herbarium/lib"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

type ModCard struct {
	*gtk.FlowBoxChild
	ModEntry  *lib.ModEntry
	CheckBtn  *gtk.CheckButton
	Label     *gtk.Label
	Container *gtk.Box
	Picture   *gtk.Picture
	Video     *gtk.Video
}

func NewModCard(
	app *HerbariumApp,
	mod *lib.ModEntry,
	cfgPath, dbPath string,
	onToggle func(),
) *ModCard {
	vbox := gtk.NewBox(gtk.OrientationVertical, 4)
	vbox.SetSpacing(16)
	vbox.SetHAlign(gtk.AlignCenter)
	vbox.SetVAlign(gtk.AlignStart)
	vbox.SetSizeRequest(200, -1)

	imageOverlay := gtk.NewOverlay()
	imageOverlay.SetVAlign(gtk.AlignCenter)
	imageOverlay.SetHAlign(gtk.AlignCenter)
	imageOverlay.SetSizeRequest(200, 200)

	container := gtk.NewBox(gtk.OrientationHorizontal, 0)
	container.SetHAlign(gtk.AlignCenter)
	container.SetVAlign(gtk.AlignCenter)
	container.SetSizeRequest(200, 200)
	imageOverlay.SetChild(container)

	picture := gtk.NewPicture()
	picture.SetSizeRequest(200, 200)
	picture.SetContentFit(gtk.ContentFitFill)
	container.Append(picture)

	check := gtk.NewCheckButton()
	check.AddCSSClass("selection-mode")
	check.SetHAlign(gtk.AlignEnd)
	check.SetVAlign(gtk.AlignStart)
	check.SetMarginEnd(4)
	check.SetMarginTop(4)
	check.SetActive(mod.Enabled)
	imageOverlay.AddOverlay(check)

	label := gtk.NewLabel(mod.Name)
	label.AddCSSClass("heading")
	label.AddCSSClass("title-2")
	label.SetHExpand(false)
	label.SetHAlign(gtk.AlignCenter)
	label.SetVAlign(gtk.AlignCenter)
	label.SetWrap(true)
	label.SetWrapMode(pango.WrapWordChar)

	vbox.Append(imageOverlay)
	vbox.Append(label)

	check.ConnectToggled(func() {
		enabled := check.Active()
		oldEnabled := mod.Enabled
		mod.Enabled = enabled
		if err := lib.ToggleEnabled(enabled, mod.Folder); err != nil {
			mod.Enabled = oldEnabled
			check.SetActive(oldEnabled)
			return
		}
		if onToggle != nil {
			onToggle()
		}
	})

	click := gtk.NewGestureClick()
	click.ConnectReleased(func(_ int, _ float64, _ float64) {
		check.SetActive(!check.Active())
	})
	imageOverlay.AddController(click)

	child := gtk.NewFlowBoxChild()
	clamp := adw.NewClamp()
	clamp.SetOverflow(gtk.OverflowHidden)
	clamp.AddCSSClass("card")
	clamp.SetMaximumSize(200)
	clamp.SetChild(vbox)
	child.SetChild(clamp)

	card := &ModCard{
		FlowBoxChild: child,
		ModEntry:     mod,
		CheckBtn:     check,
		Label:        label,
		Container:    container,
		Picture:      picture,
	}

	go card.GetPoster(app)

	return card
}

func (card *ModCard) GetPoster(app *HerbariumApp) bool {
	path, err := lib.GetOrDownloadCover(app.XDGName, *card.ModEntry)
	if err != nil {
		return true
	}

	file, err := os.Open(path)
	if err != nil {
		return true
	}
	defer file.Close()

	buffer := make([]byte, 6)
	_, err = file.Read(buffer)
	isGIF := err == nil && bytes.HasPrefix(buffer, []byte("GIF89a")) ||
		bytes.HasPrefix(buffer, []byte("GIF87a"))

	glib.IdleAdd(func() {
		if isGIF {
			card.Container.Remove(card.Picture)

			video := gtk.NewVideo()
			video.SetSizeRequest(200, 200)
			video.SetLoop(true)
			video.SetAutoplay(true)
			video.SetCanFocus(false)
			video.SetSensitive(false)

			mediaFile := gio.NewFileForPath(path)
			video.SetFile(mediaFile)
			card.Container.Append(video)
			card.Video = video
		} else {
			card.Picture.SetFilename(path)
			card.Picture.SetContentFit(gtk.ContentFitFill)
		}
	})
	return false
}
