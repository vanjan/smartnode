package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Creates a new page for the monitoring / stats settings
func createSettingStatsPage(home *settingsHome) *page {

    content := createSettingStatsContent(home)

    return newPage(
        home.homePage, 
        "settings-stats", 
        "Monitoring / Stats",
        "Select this to configure the monitoring and statistics gathering parts of the Smartnode, such as Grafana and Prometheus.",
        content,
    )

}


// Creates the content for the monitoring / stats settings page
func createSettingStatsContent(home *settingsHome) tview.Primitive {

    layout := newStandardLayout()

    // PLACEHOLDER
    paramDescriptions := []string{
        "The Execution client you'd like to use. Probably have to describe each one when you open this dropdown and hover over them.",
        "Select this if you have an external Execution client that you want the Smartnode to use, instead of managing its own (\"Hybrid Mode\").",
        "Enter Geth's cache size, in MB.",
    }

    // Create the settings form
    form := tview.NewForm()
	a := tview.NewDropDown().
		SetLabel("Client").
		SetOptions([]string{"Geth", "Infura", "Pocket", "Custom"}, nil)
	a.SetFocusFunc(func() {
		layout.descriptionBox.SetText(paramDescriptions[0])
	})
	form.AddFormItem(a)

	b := tview.NewCheckbox().
		SetLabel("Externally managed?")
	b.SetFocusFunc(func() {
		layout.descriptionBox.SetText(paramDescriptions[1])
	})
	form.AddFormItem(b)
    
	c := tview.NewInputField().
		SetLabel("Geth Cache (MB)").
		SetText("1024")
	c.SetFocusFunc(func() {
		layout.descriptionBox.SetText(paramDescriptions[2])
	})
	form.AddFormItem(c)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			home.md.setPage(home.homePage)
			return nil
		}
		return event
	})

    // Make it the content of the layout and set the default description text
    layout.setContent(form, form.Box, "Execution Client (Eth1) Settings")
    layout.descriptionBox.SetText(paramDescriptions[0])

    // Make the footer
    footer, height := createSettingFooter()
    layout.setFooter(footer, height)

    // Return the standard layout's grid
    return layout.grid

}