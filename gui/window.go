package main

import (
	"herbarium/lib"
	"fmt"
	"sort"
	"strings"
	"time"
	"unsafe"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type HerbariumWindow struct {
	Window             *adw.ApplicationWindow
	NavView            *adw.NavigationView
	ToolbarView        *adw.ToolbarView
	Header             *adw.HeaderBar
	MainBox            *gtk.Box
	ScrolledWindow     *gtk.ScrolledWindow
	FlowBox            *gtk.FlowBox
	StatsLabel         *gtk.Label
	SearchEntry        *gtk.SearchEntry
	StateDropdown      *gtk.DropDown
	TimeDropdown       *gtk.DropDown
	SelectAllBtn       *gtk.Button
	DeselectAllBtn     *gtk.Button
	LaunchButton       *gtk.Button
	Spinner            *gtk.Spinner
	SearchBar          *gtk.SearchBar
	SearchToggle       *gtk.ToggleButton
	SearchClamp        *adw.Clamp
	SearchBox          *gtk.Box
	StateList          *gtk.StringList
	TimeList           *gtk.StringList
	ButtonBox          *gtk.Box
	BottomBox          *gtk.Box
	ModCards           map[string]*ModCard
	AllModIndices      []string
	FilteredModIndices []string
	FilterTimeout      glib.SourceHandle
	PendingFilter      bool
	CfgPath, DbPath    string
}

func NewHerbariumWindow(app *HerbariumApp) *HerbariumWindow {
	win := adw.NewApplicationWindow((*gtk.Application)(unsafe.Pointer(app.App)))
	win.SetTitle(lib.T_("Herbarium"))
	win.SetDefaultSize(1200, 900)

	mw := &HerbariumWindow{
		Window:             win,
		ModCards:           make(map[string]*ModCard),
		AllModIndices:      make([]string, 0),
		FilteredModIndices: make([]string, 0),
		CfgPath:            "config.yaml",
		DbPath:             "mods_db.yaml",
	}

	mw.createWidgets()
	mw.setupUI()
	mw.loadMods(app)
	mw.connectSignals(app)
	mw.updateFilter()

	return mw
}

func (mw *HerbariumWindow) createWidgets() {
	mw.NavView = adw.NewNavigationView()
	mw.ToolbarView = adw.NewToolbarView()
	mw.Header = adw.NewHeaderBar()
	mw.MainBox = gtk.NewBox(gtk.OrientationVertical, 0)
	mw.ScrolledWindow = gtk.NewScrolledWindow()
	mw.FlowBox = gtk.NewFlowBox()
	mw.StatsLabel = gtk.NewLabel("")
	mw.SearchEntry = gtk.NewSearchEntry()
	mw.StateDropdown = gtk.NewDropDown(nil, nil)
	mw.TimeDropdown = gtk.NewDropDown(nil, nil)
	mw.SelectAllBtn = gtk.NewButton()
	mw.DeselectAllBtn = gtk.NewButton()
	mw.LaunchButton = gtk.NewButtonWithLabel(lib.T_("Launch Everlasting Summer"))
	mw.Spinner = gtk.NewSpinner()
	mw.SearchBar = gtk.NewSearchBar()
	mw.SearchToggle = gtk.NewToggleButton()
	mw.SearchClamp = adw.NewClamp()
	mw.SearchBox = gtk.NewBox(gtk.OrientationHorizontal, 0)
	mw.ButtonBox = gtk.NewBox(gtk.OrientationHorizontal, 8)
	mw.BottomBox = gtk.NewBox(gtk.OrientationHorizontal, 0)

	stateItems := []string{
		lib.T_("All states"),
		lib.T_("Enabled"),
		lib.T_("Disabled"),
	}
	mw.StateList = gtk.NewStringList(stateItems)

	timeItems := []string{
		lib.T_("Newest first"),
		lib.T_("Oldest first"),
		lib.T_("A to Z"),
		lib.T_("Z to A"),
		lib.T_("Today"),
		lib.T_("This week"),
		lib.T_("This month"),
		lib.T_("Last 3 months"),
		lib.T_("This year"),
	}
	mw.TimeList = gtk.NewStringList(timeItems)
}

func (mw *HerbariumWindow) setupUI() {
	mw.Window.SetContent(mw.NavView)

	page := adw.NewNavigationPage(mw.ToolbarView, lib.T_("Herbarium"))
	mw.NavView.Add(page)

	mw.ToolbarView.AddTopBar(mw.Header)
	mw.ToolbarView.SetTopBarStyle(adw.ToolbarFlat)

	mw.MainBox.SetHExpand(true)
	mw.MainBox.SetVExpand(true)

	mw.setupSearchBar()
	mw.setupFlowBox()
	mw.setupBottomBar()

	mw.ScrolledWindow.SetPolicy(gtk.PolicyAutomatic, gtk.PolicyAutomatic)
	mw.ScrolledWindow.SetHExpand(true)
	mw.ScrolledWindow.SetVExpand(true)
	mw.ScrolledWindow.SetChild(mw.FlowBox)

	mw.MainBox.Append(mw.SearchBar)
	mw.MainBox.Append(mw.ScrolledWindow)
	mw.ToolbarView.SetContent(mw.MainBox)
}

func (mw *HerbariumWindow) setupSearchBar() {
	mw.SearchBar.SetKeyCaptureWidget(mw.Window)
	mw.SearchBar.SetShowCloseButton(true)

	mw.SearchClamp.SetMaximumSize(480)

	mw.SearchBox.AddCSSClass("linked")

	mw.SearchEntry.SetHExpand(true)
	mw.SearchEntry.SetPlaceholderText(lib.T_("Search mods..."))
	mw.SearchBox.Append(mw.SearchEntry)

	mw.StateDropdown.SetModel(&mw.StateList.ListModel)
	mw.SearchBox.Append(mw.StateDropdown)

	mw.TimeDropdown.SetModel(&mw.TimeList.ListModel)
	mw.TimeDropdown.SetSelected(0)
	mw.SearchBox.Append(mw.TimeDropdown)

	mw.SearchClamp.SetChild(mw.SearchBox)
	mw.SearchBar.SetChild(mw.SearchClamp)
	mw.SearchBar.ConnectEntry(mw.SearchEntry)

	mw.SearchToggle.SetIconName("system-search-symbolic")
	mw.SearchBar.NotifyProperty("search-mode-enabled", func() {
		mw.SearchToggle.SetActive(mw.SearchBar.SearchMode())
	})
	mw.SearchToggle.ConnectClicked(func() {
		mw.SearchBar.SetSearchMode(!mw.SearchBar.SearchMode())
	})
	mw.Header.PackStart(mw.SearchToggle)

	mw.SelectAllBtn.SetIconName("object-select-symbolic")
	mw.SelectAllBtn.SetTooltipText(lib.T_("Select all"))
	mw.Header.PackStart(mw.SelectAllBtn)

	mw.DeselectAllBtn.SetIconName("list-remove-symbolic")
	mw.DeselectAllBtn.SetTooltipText(lib.T_("Clear selection"))
	mw.Header.PackStart(mw.DeselectAllBtn)

	mw.Header.PackEnd(mw.StatsLabel)
}

func (mw *HerbariumWindow) setupFlowBox() {
	mw.FlowBox.SetHomogeneous(true)
	mw.FlowBox.SetRowSpacing(12)
	mw.FlowBox.SetColumnSpacing(12)
	mw.FlowBox.SetHAlign(gtk.AlignCenter)
	mw.FlowBox.SetVAlign(gtk.AlignStart)
	mw.FlowBox.SetFocusable(true)
	mw.FlowBox.SetSelectionMode(gtk.SelectionNone)
}

func (mw *HerbariumWindow) setupBottomBar() {
	mw.BottomBox.SetHAlign(gtk.AlignCenter)
	mw.BottomBox.AddCSSClass("toolbar")

	mw.LaunchButton.AddCSSClass("suggested-action")
	mw.LaunchButton.SetSensitive(false)
	mw.LaunchButton.SetHAlign(gtk.AlignCenter)
	mw.LaunchButton.SetVAlign(gtk.AlignCenter)

	mw.Spinner.SetSizeRequest(16, 16)
	mw.Spinner.SetHAlign(gtk.AlignCenter)
	mw.Spinner.SetVAlign(gtk.AlignCenter)
	mw.Spinner.SetVisible(false)

	mw.ButtonBox.Append(mw.LaunchButton)
	mw.ButtonBox.Append(mw.Spinner)
	mw.BottomBox.Append(mw.ButtonBox)

	mw.ToolbarView.AddBottomBar(mw.BottomBox)
}

func (mw *HerbariumWindow) loadMods(app *HerbariumApp) {
	cfg, _ := lib.EnsureConfig()
	db, _ := lib.EnsureModsDB()
	lib.ScanAndUpdate(cfg, db)

	for i := range db.Mods {
		mod := &db.Mods[i]
		card := NewModCard(app, mod, mw.CfgPath, mw.DbPath, func() { mw.updateStats() })
		modID := mod.Folder
		mw.ModCards[modID] = card
		mw.AllModIndices = append(mw.AllModIndices, modID)
	}
}

func (mw *HerbariumWindow) connectSignals(app *HerbariumApp) {
	mw.SearchEntry.ConnectSearchChanged(func() {
		mw.scheduleFilterUpdate()
	})

	mw.StateDropdown.NotifyProperty("selected", func() {
		mw.scheduleFilterUpdate()
	})

	mw.TimeDropdown.NotifyProperty("selected", func() {
		mw.scheduleFilterUpdate()
	})

	mw.SelectAllBtn.ConnectClicked(func() {
		anyChanged := false
		for _, modID := range mw.FilteredModIndices {
			card := mw.ModCards[modID]
			if card != nil && !card.ModEntry.Enabled {
				card.CheckBtn.SetActive(true)
				anyChanged = true
			}
		}
		if anyChanged {
			mw.updateStats()
		}
	})

	mw.DeselectAllBtn.ConnectClicked(func() {
		anyChanged := false
		for _, modID := range mw.FilteredModIndices {
			card := mw.ModCards[modID]
			if card != nil && card.ModEntry.Enabled {
				card.CheckBtn.SetActive(false)
				anyChanged = true
			}
		}
		if anyChanged {
			mw.updateStats()
		}
	})

	mw.LaunchButton.ConnectClicked(func() {
		mw.LaunchButton.SetSensitive(false)
		mw.Spinner.SetVisible(true)
		mw.Spinner.Start()

		go func() {
			cfg, _ := lib.EnsureConfig()
			db, _ := lib.EnsureModsDB()
			lib.ScanAndUpdate(cfg, db)

			lib.LaunchWithMods(cfg, db)

			glib.IdleAdd(func() {
				mw.Spinner.Stop()
				mw.Spinner.SetVisible(false)
				mw.LaunchButton.SetSensitive(true)
			})
		}()
	})

	mw.LaunchButton.SetSensitive(true)
}

func (mw *HerbariumWindow) scheduleFilterUpdate() {
	if mw.FilterTimeout > 0 {
		glib.SourceRemove(mw.FilterTimeout)
	}
	mw.PendingFilter = true
	mw.FilterTimeout = glib.TimeoutAdd(250, func() bool {
		if mw.PendingFilter {
			mw.updateFilter()
			mw.PendingFilter = false
		}
		mw.FilterTimeout = 0
		return false
	})
}

func (mw *HerbariumWindow) updateFilter() {
	searchText := mw.SearchEntry.Text()
	stateFilter := mw.StateDropdown.Selected()
	timeFilter := mw.TimeDropdown.Selected()

	mw.FilteredModIndices = mw.FilteredModIndices[:0]

	modsToFilter := make([]struct {
		id   string
		card *ModCard
		time time.Time
		name string
	}, 0, len(mw.AllModIndices))

	for _, modID := range mw.AllModIndices {
		card := mw.ModCards[modID]
		if card == nil {
			continue
		}

		modsToFilter = append(modsToFilter, struct {
			id   string
			card *ModCard
			time time.Time
			name string
		}{
			id:   modID,
			card: card,
			time: card.ModEntry.DiscoveredAt,
			name: strings.ToLower(card.ModEntry.Name),
		})
	}

	switch timeFilter {
	case 0:
		sort.Slice(modsToFilter, func(i, j int) bool {
			return modsToFilter[i].time.After(modsToFilter[j].time)
		})
	case 1:
		sort.Slice(modsToFilter, func(i, j int) bool {
			return modsToFilter[i].time.Before(modsToFilter[j].time)
		})
	case 2:
		sort.Slice(modsToFilter, func(i, j int) bool {
			return modsToFilter[i].name < modsToFilter[j].name
		})
	case 3:
		sort.Slice(modsToFilter, func(i, j int) bool {
			return modsToFilter[i].name > modsToFilter[j].name
		})
	case 4:
		sort.Slice(modsToFilter, func(i, j int) bool {
			return modsToFilter[i].name < modsToFilter[j].name
		})
	default:
		sort.Slice(modsToFilter, func(i, j int) bool {
			return modsToFilter[i].time.After(modsToFilter[j].time)
		})
	}

	now := time.Now()
	for _, item := range modsToFilter {
		mod := item.card.ModEntry

		matchesSearch := true
		if searchText != "" {
			matchesSearch = strings.Contains(strings.ToLower(mod.Name), strings.ToLower(searchText))
		}

		matchesState := true
		switch stateFilter {
		case 1:
			matchesState = mod.Enabled
		case 2:
			matchesState = !mod.Enabled
		}

		matchesTime := true
		switch timeFilter {
		case 0, 1, 2, 3:
			matchesTime = true
		case 4:
			matchesTime = mod.DiscoveredAt.After(now.Add(-24 * time.Hour))
		case 5:
			matchesTime = mod.DiscoveredAt.After(now.Add(-7 * 24 * time.Hour))
		case 6:
			matchesTime = mod.DiscoveredAt.After(now.Add(-30 * 24 * time.Hour))
		case 7:
			matchesTime = mod.DiscoveredAt.After(now.Add(-90 * 24 * time.Hour))
		case 8:
			matchesTime = mod.DiscoveredAt.After(now.Add(-365 * 24 * time.Hour))
		}

		if matchesSearch && matchesState && matchesTime {
			mw.FilteredModIndices = append(mw.FilteredModIndices, item.id)
		}
	}

	mw.FlowBox.RemoveAll()
	for _, modID := range mw.FilteredModIndices {
		card := mw.ModCards[modID]
		if card != nil {
			mw.FlowBox.Append(card.FlowBoxChild)
		}
	}

	mw.updateStats()
}

func (mw *HerbariumWindow) updateStats() {
	total := len(mw.AllModIndices)
	enabled := 0
	disabled := 0

	for _, modID := range mw.AllModIndices {
		card := mw.ModCards[modID]
		if card != nil && card.ModEntry.Enabled {
			enabled++
		} else {
			disabled++
		}
	}

	statsText := fmt.Sprintf("%s: %d | %s: %d | %s: %d",
		lib.T_("Total"), total,
		lib.T_("Enabled"), enabled,
		lib.T_("Disabled"), disabled)
	mw.StatsLabel.SetText(statsText)
}
