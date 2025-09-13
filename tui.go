package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// runInteractive launches a tview-based TUI for editing options and triggering actions.
func runInteractive(o options) error {
    app := tview.NewApplication()

    pages := tview.NewPages()

    var cachedMeta *clientMeta
    depCatalog := make(map[string]depOption) // id -> dep info

	// State: selected dependency IDs
	selectedDeps := make(map[string]bool)
	if strings.TrimSpace(o.dependencies) != "" {
		for _, id := range strings.Split(o.dependencies, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				selectedDeps[id] = true
			}
		}
	}

	// Widgets
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Spring Initializr CLI ")
	form.SetButtonsAlign(tview.AlignRight)
	form.SetItemPadding(1)

	// DropDown helpers
	projectTypes := []string{"maven-project", "gradle-project", "gradle-build"}
	languages := []string{"java", "kotlin", "groovy"}
	packagings := []string{"jar", "war"}

	ddProjectType := tview.NewDropDown().SetOptions(projectTypes, nil)
	ddLanguage := tview.NewDropDown().SetOptions(languages, nil)
	ddPackaging := tview.NewDropDown().SetOptions(packagings, nil)

	// Set initial selections
	setDropDownValue(ddProjectType, projectTypes, o.projectType)
	setDropDownValue(ddLanguage, languages, o.language)
	setDropDownValue(ddPackaging, packagings, o.packaging)

	inBootVersion := tview.NewInputField().SetText(o.bootVersion)
	inGroupID := tview.NewInputField().SetText(o.groupID)
	inArtifactID := tview.NewInputField().SetText(o.artifactID)
	inName := tview.NewInputField().SetText(o.name)
	inDescription := tview.NewInputField().SetText(o.description)
	inPackageName := tview.NewInputField().SetText(o.packageName)
	inJavaVersion := tview.NewInputField().SetText(o.javaVersion)
	inBaseDir := tview.NewInputField().SetText(o.baseDir)
	inOutput := tview.NewInputField().SetText(o.output)
	inBaseURL := tview.NewInputField().SetText(o.baseURL)

	// Small helper to pull current form values into options
	readOptions := func() options {
		// DropDowns
		if _, v := ddProjectType.GetCurrentOption(); v != "" {
			o.projectType = v
		}
		if _, v := ddLanguage.GetCurrentOption(); v != "" {
			o.language = v
		}
		if _, v := ddPackaging.GetCurrentOption(); v != "" {
			o.packaging = v
		}
		// Inputs
		o.bootVersion = strings.TrimSpace(inBootVersion.GetText())
		o.groupID = strings.TrimSpace(inGroupID.GetText())
		o.artifactID = strings.TrimSpace(inArtifactID.GetText())
		o.name = strings.TrimSpace(inName.GetText())
		o.description = strings.TrimSpace(inDescription.GetText())
		o.packageName = sanitizePackage(strings.TrimSpace(inPackageName.GetText()))
		o.javaVersion = strings.TrimSpace(inJavaVersion.GetText())
		o.baseDir = strings.TrimSpace(inBaseDir.GetText())
		o.output = strings.TrimSpace(inOutput.GetText())
		o.baseURL = strings.TrimSpace(inBaseURL.GetText())
		// Dependencies
		o.dependencies = joinSelected(selectedDeps)

		// Derived defaults if empty
		if o.baseDir == "" {
			o.baseDir = o.artifactID
		}
		if o.packageName == "" {
			o.packageName = sanitizePackage(o.groupID + "." + o.artifactID)
		}
		if o.output == "" {
			o.output = o.artifactID + ".zip"
		}
		return o
	}

	// Build form items
	form.AddFormItem(labeled(ddProjectType, "Project Type"))
	form.AddFormItem(labeled(ddLanguage, "Language"))
	form.AddInputField("Boot Version", o.bootVersion, 0, nil, nil)
	form.AddInputField("Group ID", o.groupID, 0, nil, nil)
	form.AddInputField("Artifact ID", o.artifactID, 0, nil, nil)
	form.AddInputField("Name", o.name, 0, nil, nil)
	form.AddInputField("Description", o.description, 0, nil, nil)
	form.AddFormItem(labeled(ddPackaging, "Packaging"))
	form.AddInputField("Java Version", o.javaVersion, 0, nil, nil)
	form.AddInputField("Package Name", o.packageName, 0, nil, nil)
	form.AddInputField("Base Dir", o.baseDir, 0, nil, nil)
	form.AddInputField("Output Zip", o.output, 0, nil, nil)
	form.AddInputField("Base URL", o.baseURL, 0, nil, nil)

	// Hook form items to variables so readOptions sees updated values
	form.GetFormItem(2).(*tview.InputField).SetChangedFunc(func(t string) { inBootVersion.SetText(t) })
	form.GetFormItem(3).(*tview.InputField).SetChangedFunc(func(t string) { inGroupID.SetText(t) })
	form.GetFormItem(4).(*tview.InputField).SetChangedFunc(func(t string) { inArtifactID.SetText(t) })
	form.GetFormItem(5).(*tview.InputField).SetChangedFunc(func(t string) { inName.SetText(t) })
	form.GetFormItem(6).(*tview.InputField).SetChangedFunc(func(t string) { inDescription.SetText(t) })
	form.GetFormItem(8).(*tview.InputField).SetChangedFunc(func(t string) { inJavaVersion.SetText(t) })
	form.GetFormItem(9).(*tview.InputField).SetChangedFunc(func(t string) { inPackageName.SetText(t) })
	form.GetFormItem(10).(*tview.InputField).SetChangedFunc(func(t string) { inBaseDir.SetText(t) })
	form.GetFormItem(11).(*tview.InputField).SetChangedFunc(func(t string) { inOutput.SetText(t) })
	form.GetFormItem(12).(*tview.InputField).SetChangedFunc(func(t string) { inBaseURL.SetText(t) })

	// Buttons
    var postRun func() error // set when Download/Extract is chosen

    form.AddButton("Select Dependencies", func() {
        // fetch and show selector
        curr := readOptions()
        showDepsSelector(app, pages, curr.baseURL, curr.timeout, selectedDeps, depCatalog)
    })
    form.AddButton("Show Selected", func() {
        lines := selectedDisplayLines(selectedDeps, depCatalog)
        var body string
        if len(lines) == 0 {
            body = "(none)"
        } else {
            body = strings.Join(lines, "\n")
        }
        showTextModal(app, pages, "Selected dependencies", body+"\n\nEsc/Enter to close.", nil)
    })
	form.AddButton("Choose Boot Version", func() {
		curr := readOptions()
		ensureMeta := func() (*clientMeta, error) {
			if cachedMeta != nil {
				return cachedMeta, nil
			}
			m, err := fetchClientMetadata(curr.baseURL, curr.timeout)
			if err == nil {
				cachedMeta = m
				// If metadata provides canonical options, update dropdowns
				if len(m.Types) > 0 {
					ddProjectType.SetOptions(m.Types, nil)
					setDropDownValue(ddProjectType, m.Types, o.projectType)
				}
				if len(m.Languages) > 0 {
					ddLanguage.SetOptions(m.Languages, nil)
					setDropDownValue(ddLanguage, m.Languages, o.language)
				}
				if len(m.Packagings) > 0 {
					ddPackaging.SetOptions(m.Packagings, nil)
					setDropDownValue(ddPackaging, m.Packagings, o.packaging)
				}
			}
			return m, err
		}
		go func() {
			m, err := ensureMeta()
			app.QueueUpdateDraw(func() {
				if err != nil {
					showModal(app, pages, fmt.Sprintf("Failed to load metadata\n%v", err), 6*time.Second, nil)
					return
				}
				if len(m.BootVersions) == 0 {
					showModal(app, pages, "No boot versions from metadata", 4*time.Second, nil)
					return
				}
				showStringPicker(app, pages, "Select Boot Version", m.BootVersions, func(sel string) {
					inBootVersion.SetText(sel)
					// Update backing form item too
					form.GetFormItem(2).(*tview.InputField).SetText(sel)
				})
			})
		}()
	})
	form.AddButton("Choose Java Version", func() {
		curr := readOptions()
		go func() {
			var m *clientMeta
			var err error
			if cachedMeta != nil {
				m = cachedMeta
			} else {
				m, err = fetchClientMetadata(curr.baseURL, curr.timeout)
				if err == nil {
					cachedMeta = m
				}
			}
			app.QueueUpdateDraw(func() {
				if err != nil {
					showModal(app, pages, fmt.Sprintf("Failed to load metadata\n%v", err), 6*time.Second, nil)
					return
				}
				if len(m.JavaVersions) == 0 {
					showModal(app, pages, "No java versions from metadata", 4*time.Second, nil)
					return
				}
				showStringPicker(app, pages, "Select Java Version", m.JavaVersions, func(sel string) {
					inJavaVersion.SetText(sel)
					form.GetFormItem(8).(*tview.InputField).SetText(sel)
				})
			})
		}()
	})
	form.AddButton("Show URL", func() {
		curr := readOptions()
		if u, err := buildURL(curr); err != nil {
			showModal(app, pages, fmt.Sprintf("Error: %v", err), 60*time.Second, nil)
		} else {
			showTextModal(app, pages, "Generated URL", u+"\n\nPress Esc or Enter to close.", nil)
		}
	})
	form.AddButton("Download", func() {
		curr := readOptions()
		curr.dryRun = false
		curr.extract = false
		curr.interactive = false
		postRun = func() error { return run(curr) }
		app.Stop()
	})
	form.AddButton("Download+Extract", func() {
		curr := readOptions()
		curr.dryRun = false
		curr.extract = true
		curr.interactive = false
		postRun = func() error { return run(curr) }
		app.Stop()
	})
	form.AddButton("Quit", func() { app.Stop() })

	// Layout
	frame := tview.NewFrame(form).
		SetBorders(0, 0, 0, 0, 1, 1).
		AddText("Tab/Shift+Tab to move, Enter to activate.", true, tview.AlignLeft, tview.Styles.SecondaryTextColor).
		AddText("Dependencies: Enter/Space toggle, 'd' to done. Use filter.", true, tview.AlignLeft, tview.Styles.SecondaryTextColor)

	pages.AddPage("main", frame, true, true)

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		return err
	}
	if postRun != nil {
		return postRun()
	}
	return nil
}

// labeled wraps a form item with a label since DropDown lacks SetLabel in older tview.
func labeled(item tview.FormItem, label string) tview.FormItem {
	switch it := item.(type) {
	case *tview.DropDown:
		it.SetLabel(label + ": ")
	default:
		// pass-through
	}
	return item
}

func setDropDownValue(dd *tview.DropDown, options []string, val string) {
	idx := 0
	for i, v := range options {
		if strings.EqualFold(v, val) {
			idx = i
			break
		}
	}
	dd.SetCurrentOption(idx)
}

func joinSelected(m map[string]bool) string {
    ids := make([]string, 0, len(m))
    for id, ok := range m {
        if ok {
            ids = append(ids, id)
        }
    }
    sort.Strings(ids)
    return strings.Join(ids, ",")
}

func selectedIDs(m map[string]bool) []string {
    ids := make([]string, 0, len(m))
    for id, ok := range m {
        if ok {
            ids = append(ids, id)
        }
    }
    sort.Strings(ids)
    return ids
}

// selectedDisplayLines returns lines formatted as "Name (ID) [Group]" if available.
func selectedDisplayLines(selected map[string]bool, catalog map[string]depOption) []string {
    ids := selectedIDs(selected)
    if len(ids) == 0 {
        return nil
    }
    out := make([]string, 0, len(ids))
    for _, id := range ids {
        if d, ok := catalog[id]; ok {
            name := d.Name
            if name == "" {
                name = id
            }
            if d.Group != "" {
                out = append(out, fmt.Sprintf("%s (%s) [%s]", name, id, d.Group))
            } else {
                out = append(out, fmt.Sprintf("%s (%s)", name, id))
            }
        } else {
            out = append(out, id)
        }
    }
    return out
}

type depOption struct {
	ID    string
	Name  string
	Group string
}

// clientMeta holds selected lists parsed from /metadata/client
type clientMeta struct {
	Types        []string
	Languages    []string
	Packagings   []string
	JavaVersions []string
	BootVersions []string
}

// fetchClientMetadata returns lists from /metadata/client for dropdowns and pickers.
func fetchClientMetadata(baseURL string, timeout int) (*clientMeta, error) {
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	base := strings.TrimRight(baseURL, "/")
	req, _ := http.NewRequest(http.MethodGet, base+"/metadata/client", nil)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %s", resp.Status)
	}

	var data map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	extractIDs := func(key string) []string {
		raw, ok := data[key]
		if !ok {
			return nil
		}
		var sec struct {
			Values []struct {
				ID string `json:"id"`
			} `json:"values"`
		}
		if err := json.Unmarshal(raw, &sec); err != nil {
			return nil
		}
		out := make([]string, 0, len(sec.Values))
		for _, v := range sec.Values {
			if v.ID != "" {
				out = append(out, v.ID)
			}
		}
		return out
	}
	m := &clientMeta{
		Types:        extractIDs("type"),
		Languages:    extractIDs("language"),
		Packagings:   extractIDs("packaging"),
		JavaVersions: extractIDs("javaVersion"),
		BootVersions: extractIDs("bootVersion"),
	}
	return m, nil
}

func showDepsSelector(app *tview.Application, pages *tview.Pages, baseURL string, timeout int, selected map[string]bool, catalog map[string]depOption) {
	// Show loading modal while fetching
	loading := tview.NewModal().SetText("Fetching dependencies...\n(Press Esc to cancel)")
	loading.AddButtons([]string{"Cancel"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pages.RemovePage("loading")
	})
	pages.AddPage("loading", centered(loading, 0.4, 0.3), true, true)

	go func() {
		deps, err := fetchDependencies(baseURL, timeout)
		app.QueueUpdateDraw(func() {
			pages.RemovePage("loading")
			if err != nil {
				showModal(app, pages, fmt.Sprintf("Failed to fetch dependencies:\n%v", err), 8*time.Second, nil)
				return
			}
            // Sort by group then name
            sort.Slice(deps, func(i, j int) bool {
                if deps[i].Group == deps[j].Group {
                    return strings.ToLower(deps[i].Name) < strings.ToLower(deps[j].Name)
                }
                return strings.ToLower(deps[i].Group) < strings.ToLower(deps[j].Group)
            })

            // Update catalog
            for _, d := range deps {
                if d.ID != "" {
                    catalog[d.ID] = d
                }
            }

			// UI: filter input + list
			filter := tview.NewInputField().SetLabel("Filter: ")
			list := tview.NewList()
			list.SetBorder(true).SetTitle(" Select Dependencies (Enter/Space: toggle, d: done) ")

			// Helper to build visible list from filter
			filtered := make([]depOption, len(deps))
			copy(filtered, deps)
			rebuild := func() {
				q := strings.ToLower(strings.TrimSpace(filter.GetText()))
				filtered = filtered[:0]
				for _, d := range deps {
					if q == "" || strings.Contains(strings.ToLower(d.ID), q) || strings.Contains(strings.ToLower(d.Name), q) || strings.Contains(strings.ToLower(d.Group), q) {
						filtered = append(filtered, d)
					}
				}
				list.Clear()
				for _, d := range filtered {
					list.AddItem(depLabel(d, selected[d.ID]), "", 0, nil)
				}
			}
            filter.SetChangedFunc(func(text string) { rebuild() })
            // Tab/Backtab/Enter move focus from filter -> list. Esc closes.
            filter.SetDoneFunc(func(key tcell.Key) {
                switch key {
                case tcell.KeyTab, tcell.KeyEnter:
                    app.SetFocus(list)
                case tcell.KeyBacktab:
                    // stay on filter; no-op to avoid leaving dialog
                case tcell.KeyEsc:
                    pages.RemovePage("deps")
                }
            })
            filter.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
                if ev.Key() == tcell.KeyDown {
                    app.SetFocus(list)
                    return nil
                }
                return ev
            })

			rebuild()

			list.SetDoneFunc(func() {
				pages.RemovePage("deps")
				app.SetFocus(pages)
			})
            list.SetSelectedFunc(func(i int, mainText, secondaryText string, shortcut rune) {
                if i >= 0 && i < len(filtered) {
                    d := filtered[i]
                    selected[d.ID] = !selected[d.ID]
                    list.SetItemText(i, depLabel(d, selected[d.ID]), "")
                    // If newly checked, clear filter to show full list again
                    if selected[d.ID] {
                        filter.SetText("")
                        rebuild()
                        app.SetFocus(filter)
                    }
                }
            })
            list.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
                switch ev.Rune() {
                case 'd', 'D':
                    pages.RemovePage("deps")
                    return nil
                case ' ': // Space toggles
                    if i := list.GetCurrentItem(); i >= 0 && i < len(filtered) {
                        d := filtered[i]
                        selected[d.ID] = !selected[d.ID]
                        list.SetItemText(i, depLabel(d, selected[d.ID]), "")
                        if selected[d.ID] {
                            filter.SetText("")
                            rebuild()
                            app.SetFocus(filter)
                        }
                        return nil
                    }
                case '/': // jump to filter input
                    app.SetFocus(filter)
                    return nil
                }
                switch ev.Key() {
                case tcell.KeyTab, tcell.KeyBacktab:
                    app.SetFocus(filter)
                    return nil
                }
                return ev
            })

			flex := tview.NewFlex().SetDirection(tview.FlexRow)
			flex.AddItem(filter, 1, 0, true)
			flex.AddItem(list, 0, 1, false)
            help := tview.NewTextView().SetText("Tab: Filter/List  |  /: focus Filter  |  Enter/Space: toggle  |  d: done  |  Esc: close  |  Type to filter")
			help.SetTextColor(tview.Styles.SecondaryTextColor)
			flex.AddItem(help, 1, 0, false)

			pages.AddPage("deps", centered(flex, 0.8, 0.9), true, true)
			app.SetFocus(filter)
		})
	}()
}

func depLabel(d depOption, checked bool) string {
	mark := "☐"
	if checked {
		mark = "☑"
	}
	grp := d.Group
	if grp != "" {
		grp = " [" + grp + "]"
	}
	return fmt.Sprintf("%s %s (%s)%s", mark, d.Name, d.ID, grp)
}

func centered(p tview.Primitive, widthPct, heightPct float64) tview.Primitive {
	grid := tview.NewGrid().
		SetColumns(0, int(widthPct*100), 0).
		SetRows(0, int(heightPct*100), 0).
		SetBorders(false)
	grid.AddItem(p, 1, 1, 1, 1, 0, 0, true)
	return grid
}

func showModal(app *tview.Application, pages *tview.Pages, msg string, autoClose time.Duration, onClose func()) {
	modal := tview.NewModal().SetText(msg).AddButtons([]string{"OK"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pages.RemovePage("modal")
		if onClose != nil {
			onClose()
		}
	})
	pages.AddPage("modal", centered(modal, 0.5, 0.4), true, true)
	if autoClose > 0 {
		time.AfterFunc(autoClose, func() {
			app.QueueUpdateDraw(func() {
				pages.RemovePage("modal")
				if onClose != nil {
					onClose()
				}
			})
		})
	}
}

func showTextModal(app *tview.Application, pages *tview.Pages, title, text string, onClose func()) {
	tv := tview.NewTextView().SetText(text).SetScrollable(true)
	tv.SetBorder(true).SetTitle(" " + title + " ")
	frame := tview.NewFrame(tv).
		AddText("Esc/Enter to close", true, tview.AlignCenter, tview.Styles.SecondaryTextColor)
	pages.AddPage("text", centered(frame, 0.8, 0.8), true, true)
	tv.SetDoneFunc(func(key tcell.Key) {
		pages.RemovePage("text")
		if onClose != nil {
			onClose()
		}
	})
	tv.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEsc, tcell.KeyEnter:
			pages.RemovePage("text")
			if onClose != nil {
				onClose()
			}
			return nil
		}
		return ev
	})
}

// showStringPicker shows a simple list picker for selecting a string from items.
func showStringPicker(app *tview.Application, pages *tview.Pages, title string, items []string, onChoose func(string)) {
    list := tview.NewList()
    list.SetBorder(true).SetTitle(" " + title + " ")
	for _, it := range items {
		val := it
		list.AddItem(val, "", 0, func() {
			if onChoose != nil {
				onChoose(val)
			}
			pages.RemovePage("picker")
		})
	}
	list.SetDoneFunc(func() {
		pages.RemovePage("picker")
	})
	pages.AddPage("picker", centered(list, 0.5, 0.7), true, true)
	app.SetFocus(list)
}

// Fetch dependencies from Initializr metadata endpoints, tolerating schema variants.
func fetchDependencies(baseURL string, timeout int) ([]depOption, error) {
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	base := strings.TrimRight(baseURL, "/")

	// Try /metadata/client first
	if deps, err := fetchFromMetadataClient(client, base+"/metadata/client"); err == nil && len(deps) > 0 {
		return deps, nil
	}
	// Fallback to /dependencies
	if deps, err := fetchFromDependencies(client, base+"/dependencies"); err == nil && len(deps) > 0 {
		return deps, nil
	} else if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("no dependencies found from %s", base)
}

func fetchFromMetadataClient(client *http.Client, url string) ([]depOption, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %s", resp.Status)
	}
	var data struct {
		Dependencies struct {
			Values []struct {
				Name   string            `json:"name"`
				Values []json.RawMessage `json:"values"`
			} `json:"values"`
		} `json:"dependencies"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	var out []depOption
	for _, grp := range data.Dependencies.Values {
		gname := grp.Name
		for _, raw := range grp.Values {
			var item struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			if err := json.Unmarshal(raw, &item); err == nil && item.ID != "" {
				out = append(out, depOption{ID: item.ID, Name: item.Name, Group: gname})
			}
		}
	}
	return out, nil
}

func fetchFromDependencies(client *http.Client, url string) ([]depOption, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %s", resp.Status)
	}

	// tolerate both {groups:[{name,values:[{id,name}]}]} and {dependencies:[{id,name,group}]}
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if graw, ok := raw["groups"]; ok {
		var groups []struct {
			Name   string `json:"name"`
			Values []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"values"`
		}
		if err := json.Unmarshal(graw, &groups); err == nil {
			var out []depOption
			for _, g := range groups {
				for _, v := range g.Values {
					out = append(out, depOption{ID: v.ID, Name: v.Name, Group: g.Name})
				}
			}
			return out, nil
		}
	}
	if draw, ok := raw["dependencies"]; ok {
		var deps []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Group string `json:"group"`
		}
		if err := json.Unmarshal(draw, &deps); err == nil {
			out := make([]depOption, 0, len(deps))
			for _, d := range deps {
				out = append(out, depOption{ID: d.ID, Name: d.Name, Group: d.Group})
			}
			return out, nil
		}
	}
	return nil, fmt.Errorf("unsupported dependencies schema")
}
