// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	appconfig "github.com/alyraffauf/switchyard/internal/config"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	coreglib "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

type launcherMetricsType struct {
	IconSize       int
	IconPadding    int
	ButtonSpacing  int
	ContentMarginH int
	ContentMarginT int
	ContentMarginB int
	BarMargin      int
	MaxColumns     uint
	MinColumns     uint
	MediumBreak    float64
	NarrowBreak    float64
	MediumColumns  uint
	NarrowColumns  uint
}

func (m launcherMetricsType) ButtonSize() int {
	return m.IconSize + m.IconPadding
}

var launcherMetrics = launcherMetricsType{
	IconSize:       128,
	IconPadding:    6,
	ButtonSpacing:  16,
	ContentMarginH: 16,
	ContentMarginT: 16,
	ContentMarginB: 8,
	BarMargin:      8,
	MaxColumns:     6,
	MinColumns:     2,
	MediumBreak:    600,
	NarrowBreak:    400,
	MediumColumns:  3,
	NarrowColumns:  2,
}

// loadBrowserIcon builds an image for the browser, preferring the live GIcon so
// GTK picks the sharpest themed size instead of blurry named-icon scaling.
func loadBrowserIcon(browser *Browser, size int) *gtk.Image {
	if browser.appInfo != nil {
		if gicon := browser.appInfo.Icon(); gicon != nil {
			image := gtk.NewImageFromGIcon(gicon)
			image.SetPixelSize(size)
			return image
		}
	}

	iconName := browser.Icon
	if iconName == "" {
		iconName = "web-browser-symbolic"
	}

	image := gtk.NewImageFromIconName(iconName)
	image.SetPixelSize(size)
	return image
}

// BrowserButtonCallbacks holds the handlers for browser button interactions.
type BrowserButtonCallbacks struct {
	OnClick      func(browser *Browser)
	OnRightClick func(btn *gtk.Button, browser *Browser)
}

func createBrowserButton(browser *Browser, showLabel bool, callbacks BrowserButtonCallbacks) *gtk.Button {
	btn := gtk.NewButton()
	btn.AddCSSClass("flat")
	btn.SetSizeRequest(launcherMetrics.ButtonSize(), launcherMetrics.ButtonSize())

	// Container inside button - icon above, optional name below
	btnBox := gtk.NewBox(gtk.OrientationVertical, 8)
	btnBox.SetHAlign(gtk.AlignCenter)
	btnBox.SetVAlign(gtk.AlignCenter)

	// Fixed-size container for icon to ensure uniform sizing
	iconBox := gtk.NewBox(gtk.OrientationVertical, 0)
	iconBox.SetSizeRequest(launcherMetrics.IconSize, launcherMetrics.IconSize)
	iconBox.SetHAlign(gtk.AlignCenter)
	iconBox.SetVAlign(gtk.AlignCenter)

	icon := loadBrowserIcon(browser, launcherMetrics.IconSize)
	icon.SetHAlign(gtk.AlignCenter)
	icon.SetVAlign(gtk.AlignCenter)
	iconBox.Append(icon)

	btnBox.Append(iconBox)

	// Set accessible label for screen readers (always, regardless of visual label)
	labelValue := coreglib.NewValue(browser.Name)
	btn.UpdateProperty([]gtk.AccessibleProperty{gtk.AccessiblePropertyLabel}, []coreglib.Value{*labelValue})

	if showLabel {
		label := gtk.NewLabel(browser.Name)
		label.SetEllipsize(pango.EllipsizeEnd)
		label.SetMaxWidthChars(18)
		label.SetJustify(gtk.JustifyCenter)
		label.SetLines(1)
		label.SetMarginTop(6)
		btnBox.Append(label)
	} else {
		btn.SetTooltipText(browser.Name)
	}

	btn.SetChild(btnBox)

	if callbacks.OnClick != nil {
		b := browser
		btn.ConnectClicked(func() {
			callbacks.OnClick(b)
		})
	}

	if callbacks.OnRightClick != nil {
		b := browser
		gesture := gtk.NewGestureClick()
		gesture.SetButton(gdk.BUTTON_SECONDARY)
		gesture.ConnectPressed(func(nPress int, x, y float64) {
			callbacks.OnRightClick(btn, b)
		})
		btn.AddController(gesture)
	}

	return btn
}

// LauncherFlowBox wraps a FlowBox with its launcher breakpoints.
type LauncherFlowBox struct {
	FlowBox          *gtk.FlowBox
	NarrowBreakpoint *adw.Breakpoint
	MediumBreakpoint *adw.Breakpoint
}

// Returns the flow box and breakpoints that should be added to the window.
func createLauncherFlowBox() *LauncherFlowBox {
	flowBox := gtk.NewFlowBox()
	flowBox.SetSelectionMode(gtk.SelectionSingle)
	flowBox.SetActivateOnSingleClick(false)
	flowBox.SetColumnSpacing(uint(launcherMetrics.ButtonSpacing))
	flowBox.SetRowSpacing(uint(launcherMetrics.ButtonSpacing))
	flowBox.SetMaxChildrenPerLine(launcherMetrics.MaxColumns)
	flowBox.SetMinChildrenPerLine(launcherMetrics.MinColumns)
	flowBox.SetHAlign(gtk.AlignCenter)
	flowBox.SetVAlign(gtk.AlignStart)

	// Breakpoint for narrow windows (mobile/small screens)
	narrowBreakpoint := adw.NewBreakpoint(adw.NewBreakpointConditionLength(
		adw.BreakpointConditionMaxWidth, launcherMetrics.NarrowBreak, adw.LengthUnitPx))
	narrowBreakpoint.AddSetter(flowBox, "max-children-per-line", launcherMetrics.NarrowColumns)

	// Breakpoint for medium windows
	mediumBreakpoint := adw.NewBreakpoint(adw.NewBreakpointConditionLength(
		adw.BreakpointConditionMaxWidth, launcherMetrics.MediumBreak, adw.LengthUnitPx))
	mediumBreakpoint.AddSetter(flowBox, "max-children-per-line", launcherMetrics.MediumColumns)

	return &LauncherFlowBox{
		FlowBox:          flowBox,
		NarrowBreakpoint: narrowBreakpoint,
		MediumBreakpoint: mediumBreakpoint,
	}
}

func createLauncherContentBox() *gtk.Box {
	contentBox := gtk.NewBox(gtk.OrientationVertical, 0)
	contentBox.SetMarginStart(launcherMetrics.ContentMarginH)
	contentBox.SetMarginEnd(launcherMetrics.ContentMarginH)
	contentBox.SetMarginTop(launcherMetrics.ContentMarginT)
	contentBox.SetMarginBottom(launcherMetrics.ContentMarginB)
	return contentBox
}

func createLauncherBottomBar(urlEntry *gtk.Entry, onClose func()) *gtk.Box {
	bottomBar := gtk.NewBox(gtk.OrientationHorizontal, 12)
	bottomBar.SetMarginStart(launcherMetrics.BarMargin)
	bottomBar.SetMarginEnd(launcherMetrics.BarMargin)
	bottomBar.SetMarginTop(launcherMetrics.BarMargin)
	bottomBar.SetMarginBottom(launcherMetrics.BarMargin)

	menuBtn := gtk.NewMenuButton()
	menuBtn.SetIconName("open-menu-symbolic")
	menuBtn.SetTooltipText("Main menu")
	menuBtn.AddCSSClass("flat")

	menu := gio.NewMenu()
	menu.Append("Settings", "win.settings")

	aboutSection := gio.NewMenu()
	aboutSection.Append("Donate ❤️", "win.donate")
	aboutSection.Append("About", "win.about")
	aboutSection.Append("Keyboard Shortcuts", "win.shortcuts")
	menu.AppendSection("", aboutSection)

	quitSection := gio.NewMenu()
	quitSection.Append("Quit", "win.quit")
	menu.AppendSection("", quitSection)

	menuBtn.SetMenuModel(menu)
	bottomBar.Append(menuBtn)

	leftSpacer := gtk.NewBox(gtk.OrientationHorizontal, 0)
	leftSpacer.SetHExpand(true)
	bottomBar.Append(leftSpacer)

	bottomBar.Append(urlEntry)

	rightSpacer := gtk.NewBox(gtk.OrientationHorizontal, 0)
	rightSpacer.SetHExpand(true)
	bottomBar.Append(rightSpacer)

	closeBtn := gtk.NewButton()
	closeBtn.SetIconName("window-close-symbolic")
	closeBtn.SetTooltipText("Close")
	closeBtn.AddCSSClass("circular")
	closeBtn.ConnectClicked(onClose)
	bottomBar.Append(closeBtn)

	return bottomBar
}

func createURLEntry(url string) *gtk.Entry {
	urlEntry := gtk.NewEntry()
	urlEntry.SetText(url)
	urlEntry.SetEditable(true)
	urlEntry.SetCanFocus(true)
	urlEntry.SetAlignment(0.5)
	urlEntry.SetMaxWidthChars(50)
	urlEntry.SetWidthChars(40)
	return urlEntry
}

// getLogicFromComboRow extracts the logic string from a combo row selection
func getLogicFromComboRow(logicRow *adw.ComboRow) string {
	if logicRow.Selected() == 1 {
		return "any"
	}
	return "all"
}

// saveConfigWithFlag saves config while setting the global saving flag to prevent file watcher loops
func saveConfigWithFlag(cfg *Config) {
	savingMux.Lock()
	isSaving = true
	savingMux.Unlock()
	appconfig.Save(appconfig.Path(), cfg)
	glib.TimeoutAdd(100, func() bool {
		savingMux.Lock()
		isSaving = false
		savingMux.Unlock()
		return false
	})
}

// findBrowserByID finds a browser by its desktop file ID
func findBrowserByID(browsers []*Browser, id string) *Browser {
	for _, b := range browsers {
		if b.ID == id {
			return b
		}
	}
	return nil
}

// settingsPageLayout creates the standard layout for settings pages.
// Returns the toolbar view, content box for adding widgets, and the header bar.
func settingsPageLayout(title string) (*adw.ToolbarView, *gtk.Box, *adw.HeaderBar) {
	toolbarView := adw.NewToolbarView()

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(true)
	titleLabel := gtk.NewLabel(title)
	titleLabel.AddCSSClass("title")
	header.SetTitleWidget(titleLabel)
	toolbarView.AddTopBar(header)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetVExpand(true)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

	content := gtk.NewBox(gtk.OrientationVertical, 24)
	content.SetMarginStart(24)
	content.SetMarginEnd(24)
	content.SetMarginTop(24)
	content.SetMarginBottom(24)

	clamp := adw.NewClamp()
	clamp.SetMaximumSize(600)
	clamp.SetChild(content)
	scrolled.SetChild(clamp)

	toolbarView.SetContent(scrolled)
	return toolbarView, content, header
}

// createScrolledWindow creates a standard scrolled window for list content.
func createScrolledWindow() *gtk.ScrolledWindow {
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetVExpand(true)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	return scrolled
}

// createBoxedListBox creates a ListBox with standard boxed-list styling.
func createBoxedListBox() *gtk.ListBox {
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionNone)
	listBox.AddCSSClass("boxed-list")
	return listBox
}

// createEmptyState creates a status page for empty list states.
func createEmptyState(iconName, title, description string) *adw.StatusPage {
	emptyState := adw.NewStatusPage()
	emptyState.SetIconName(iconName)
	emptyState.SetTitle(title)
	emptyState.SetDescription(description)
	emptyState.SetVExpand(true)
	return emptyState
}

// clearListBox removes all children from a ListBox.
func clearListBox(listBox *gtk.ListBox) {
	for {
		child := listBox.FirstChild()
		if child == nil {
			break
		}
		listBox.Remove(child)
	}
}

// configFileFilters creates the standard file filter list for config import/export.
func configFileFilters() *gio.ListStore {
	tomlFilter := gtk.NewFileFilter()
	tomlFilter.SetName("TOML files")
	tomlFilter.AddPattern("*.toml")

	allFilter := gtk.NewFileFilter()
	allFilter.SetName("All files")
	allFilter.AddPattern("*")

	filters := gio.NewListStore(glib.TypeObject)
	filters.Append(tomlFilter.Object)
	filters.Append(allFilter.Object)
	return filters
}
