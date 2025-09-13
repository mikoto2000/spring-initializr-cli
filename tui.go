package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "sort"
    "strings"
    "time"

    "github.com/gdamore/tcell/v2"
    "github.com/rivo/tview"
)

type depItem struct{ ID, Name string }
type depGroup struct{ Name string; Items []depItem }

// runInteractive implements a TUI using tview, allowing dependency selection from a list.
func runInteractive(initial options) error {
    app := tview.NewApplication()

    // Load dependencies asynchronously once UI starts
    groups := make([]depGroup, 0)
    selected := make(map[string]bool)
    depLoadErr := error(nil)

    // Copy initial for editing
    o := initial

    // Form fields
    form := tview.NewForm().SetHorizontal(true)
    form.AddInputField("Base URL", o.baseURL, 0, nil, func(s string) { o.baseURL = s })
    form.AddDropDown("Type", []string{"maven-project", "gradle-project", "gradle-build"}, indexOf([]string{"maven-project", "gradle-project", "gradle-build"}, o.projectType), func(option string, _ int) { o.projectType = option })
    form.AddDropDown("Language", []string{"java", "kotlin", "groovy"}, indexOf([]string{"java", "kotlin", "groovy"}, o.language), func(option string, _ int) { o.language = option })
    form.AddInputField("Boot Version", o.bootVersion, 0, nil, func(s string) { o.bootVersion = s })
    form.AddInputField("Group ID", o.groupID, 0, nil, func(s string) { o.groupID = s })
    form.AddInputField("Artifact ID", o.artifactID, 0, nil, func(s string) { o.artifactID = s })
    form.AddInputField("Name", o.name, 0, nil, func(s string) { o.name = s })
    form.AddInputField("Description", o.description, 0, nil, func(s string) { o.description = s })
    form.AddInputField("Package Name", o.packageName, 0, nil, func(s string) { o.packageName = sanitizePackage(s) })
    form.AddDropDown("Packaging", []string{"jar", "war"}, indexOf([]string{"jar", "war"}, o.packaging), func(option string, _ int) { o.packaging = option })
    form.AddInputField("Java Version", o.javaVersion, 0, nil, func(s string) { o.javaVersion = s })
    form.AddCheckbox("Extract", o.extract, func(v bool) { o.extract = v })
    form.AddCheckbox("Verbose", o.verbose, func(v bool) { o.verbose = v })

    // Buttons
    var pages *tview.Pages
    showMessage := func(title, msg string) {
        modal := tview.NewModal().SetText(msg).AddButtons([]string{"OK"}).SetDoneFunc(func(i int, l string) {
            pages.RemovePage("modal")
        })
        pages.AddPage("modal", modal, true, true)
    }

    // Dependency picker
    pickDependencies := func() {
        list := tview.NewList()
        list.SetBorder(true).SetTitle("Dependencies (Enter: toggle, d: done)")
        list.ShowSecondaryText(false)

        // helper to render
        var refresh func(filter string)
        refresh = func(filter string) {
            list.Clear()
            if len(groups) == 0 && depLoadErr == nil {
                list.AddItem("Loading...", "", 0, nil)
                return
            }
            if len(groups) == 0 && depLoadErr != nil {
                // fallback to built-in defaults
                groups = defaultDepGroups()
            }
            for _, g := range groups {
                matched := make([]depItem, 0)
                for _, d := range g.Items {
                    if filter != "" && !strings.Contains(strings.ToLower(d.Name+" "+d.ID), strings.ToLower(filter)) {
                        continue
                    }
                    matched = append(matched, d)
                }
                if len(matched) == 0 { continue }
                list.AddItem("-- "+g.Name+" --", "", 0, nil)
                for _, d := range matched {
                    label := fmt.Sprintf("[%s] %s (%s)", boolToX(selected[d.ID]), d.Name, d.ID)
                    id := d.ID
                    list.AddItem(label, "", 0, func() {
                        selected[id] = !selected[id]
                        refresh(filter)
                    })
                }
            }
        }

        input := tview.NewInputField().SetLabel("Filter: ")
        input.SetChangedFunc(func(text string) { refresh(text) })

        flex := tview.NewFlex().SetDirection(tview.FlexRow).
            AddItem(input, 1, 0, true).
            AddItem(list, 0, 1, false)

        // keybindings
        list.SetDoneFunc(func() { pages.RemovePage("picker") })
        flex.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
            if ev.Key() == tcell.KeyRune && (ev.Rune() == 'd' || ev.Rune() == 'D') {
                pages.RemovePage("picker")
                return nil
            }
            return ev
        })

        pages.AddPage("picker", tview.NewFrame(flex).SetBorders(0, 0, 0, 0, 0, 0).SetBorder(true).SetTitle("Select Dependencies"), true, true)
        refresh("")
        app.SetFocus(input)
    }

    form.AddButton("Select Dependencies", func() {
        if len(groups) == 0 && depLoadErr != nil {
            showMessage("Error", fmt.Sprintf("Failed to load dependencies: %v", depLoadErr))
            return
        }
        pickDependencies()
    })
    form.AddButton("Show URL", func() {
        // derive dependent fields
        if o.baseDir == "" { o.baseDir = o.artifactID }
        if o.packageName == "" { o.packageName = sanitizePackage(o.groupID+"."+o.artifactID) }
        o.dependencies = selectedIDsCSV(selected)
        u, err := buildURL(o)
        if err != nil {
            showMessage("Error", err.Error())
            return
        }
        showMessage("URL", u)
    })
    form.AddButton("Download", func() {
        o.dependencies = selectedIDsCSV(selected)
        go func() {
            err := run(applyAction(o, "download"))
            app.QueueUpdateDraw(func() {
                if err != nil {
                    showMessage("Error", err.Error())
                } else {
                    showMessage("Done", "Downloaded successfully")
                }
            })
        }()
    })
    form.AddButton("Download+Extract", func() {
        o.extract = true
        o.dependencies = selectedIDsCSV(selected)
        go func() {
            err := run(applyAction(o, "extract"))
            app.QueueUpdateDraw(func() {
                if err != nil {
                    showMessage("Error", err.Error())
                } else {
                    showMessage("Done", "Extracted successfully")
                }
            })
        }()
    })
    form.AddButton("Quit", func() { app.Stop() })

    form.SetBorder(true).SetTitle("Spring Initializr - TUI")

    pages = tview.NewPages().AddPage("main", form, true, true)

    // Start dependency load after UI starts
    go func(baseURL, bootVersion string) {
        list, err := fetchDependencyGroups(baseURL, bootVersion)
        if err != nil {
            depLoadErr = err
        } else {
            groups = list
        }
        // preserve previously typed CSV in initial options
        if strings.TrimSpace(initial.dependencies) != "" {
            for _, id := range strings.Split(initial.dependencies, ",") {
                selected[strings.TrimSpace(id)] = true
            }
        }
        app.QueueUpdateDraw(func() {})
    }(o.baseURL, o.bootVersion)

    if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
        return err
    }
    return nil
}

func boolToX(b bool) string { if b { return "x" } ; return " " }

func indexOf(slice []string, v string) int {
    for i, s := range slice { if s == v { return i } }
    return 0
}

func selectedIDsCSV(m map[string]bool) string {
    ids := make([]string, 0, len(m))
    for id, ok := range m { if ok { ids = append(ids, id) } }
    sort.Strings(ids)
    return strings.Join(ids, ",")
}

// fetchDependencyGroups loads dependency groups from Initializr.
// It prefers /metadata/client (which includes groups). If unavailable, falls back
// to /dependencies and wraps the flat list into a single "All" group.
func fetchDependencyGroups(baseURL, bootVersion string) ([]depGroup, error) {
    client := &http.Client{ Timeout: 15 * time.Second }
    base := strings.TrimRight(baseURL, "/")

    if groups, err := fetchGroupsFromClientMetadata(client, base); err == nil && len(groups) > 0 {
        return groups, nil
    }
    if flat, err := fetchFromDependencies(client, base, bootVersion); err == nil && len(flat) > 0 {
        g := depGroup{ Name: "All" }
        for _, f := range flat {
            g.Items = append(g.Items, depItem{ID: f.ID, Name: f.Name})
        }
        return []depGroup{ g }, nil
    }
    return nil, fmt.Errorf("failed to load dependencies from %s", baseURL)
}

func fetchFromClientMetadata(client *http.Client, base string) ([]struct{ ID, Name string }, error) {
    req, _ := http.NewRequest(http.MethodGet, base+"/metadata/client", nil)
    req.Header.Set("Accept", "application/json")
    resp, err := client.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode/100 != 2 { io.Copy(io.Discard, resp.Body); return nil, fmt.Errorf("status %s", resp.Status) }
    var data struct{
        Dependencies struct {
            Values []struct{
                Name string `json:"name"`
                Values []struct{
                    ID string `json:"id"`
                    Name string `json:"name"`
                } `json:"values"`
            } `json:"values"`
        } `json:"dependencies"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil { return nil, err }
    out := make([]struct{ID, Name string}, 0)
    for _, grp := range data.Dependencies.Values {
        for _, v := range grp.Values {
            out = append(out, struct{ID, Name string}{ID: v.ID, Name: v.Name})
        }
    }
    sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
    return out, nil
}

func fetchGroupsFromClientMetadata(client *http.Client, base string) ([]depGroup, error) {
    req, _ := http.NewRequest(http.MethodGet, base+"/metadata/client", nil)
    req.Header.Set("Accept", "application/json")
    resp, err := client.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode/100 != 2 { io.Copy(io.Discard, resp.Body); return nil, fmt.Errorf("status %s", resp.Status) }
    var data struct{
        Dependencies struct {
            Values []struct{
                Name string `json:"name"`
                Values []struct{
                    ID string `json:"id"`
                    Name string `json:"name"`
                } `json:"values"`
            } `json:"values"`
        } `json:"dependencies"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil { return nil, err }
    groups := make([]depGroup, 0, len(data.Dependencies.Values))
    for _, grp := range data.Dependencies.Values {
        g := depGroup{ Name: grp.Name }
        for _, v := range grp.Values {
            g.Items = append(g.Items, depItem{ ID: v.ID, Name: v.Name })
        }
        sort.Slice(g.Items, func(i, j int) bool { return g.Items[i].Name < g.Items[j].Name })
        groups = append(groups, g)
    }
    sort.Slice(groups, func(i, j int) bool { return groups[i].Name < groups[j].Name })
    return groups, nil
}

// defaultDepGroups provides a minimal offline list for when metadata cannot be fetched.
func defaultDepGroups() []depGroup {
    return []depGroup{
        { Name: "Web", Items: []depItem{
            {ID: "web", Name: "Spring Web"},
            {ID: "webflux", Name: "Spring Reactive Web"},
            {ID: "thymeleaf", Name: "Thymeleaf"},
            {ID: "mustache", Name: "Mustache"},
        }},
        { Name: "Data", Items: []depItem{
            {ID: "data-jpa", Name: "Spring Data JPA"},
            {ID: "data-mongodb", Name: "Spring Data MongoDB"},
            {ID: "jdbc", Name: "JDBC"},
            {ID: "r2dbc", Name: "R2DBC"},
        }},
        { Name: "Databases", Items: []depItem{
            {ID: "mysql", Name: "MySQL Driver"},
            {ID: "postgresql", Name: "PostgreSQL Driver"},
            {ID: "h2", Name: "H2 Database"},
            {ID: "flyway", Name: "Flyway"},
            {ID: "liquibase", Name: "Liquibase"},
        }},
        { Name: "Core", Items: []depItem{
            {ID: "validation", Name: "Validation"},
            {ID: "security", Name: "Spring Security"},
            {ID: "actuator", Name: "Spring Boot Actuator"},
            {ID: "lombok", Name: "Lombok"},
            {ID: "devtools", Name: "Spring Boot DevTools"},
        }},
        { Name: "Testing", Items: []depItem{
            {ID: "testcontainers", Name: "Testcontainers"},
        }},
    }
}

func fetchFromDependencies(client *http.Client, base, bootVersion string) ([]struct{ ID, Name string }, error) {
    u := base + "/dependencies"
    if strings.TrimSpace(bootVersion) != "" {
        u += "?bootVersion=" + urlQueryEscape(bootVersion)
    }
    req, _ := http.NewRequest(http.MethodGet, u, nil)
    req.Header.Set("Accept", "application/json")
    resp, err := client.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode/100 != 2 { io.Copy(io.Discard, resp.Body); return nil, fmt.Errorf("status %s", resp.Status) }
    var data struct{ Dependencies []struct{ ID, Name string } `json:"dependencies"` }
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil { return nil, err }
    out := make([]struct{ID, Name string}, 0, len(data.Dependencies))
    for _, d := range data.Dependencies { out = append(out, struct{ID, Name string}{ID: d.ID, Name: d.Name}) }
    sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
    return out, nil
}

func urlQueryEscape(s string) string {
    // local tiny escape to avoid importing net/url here; main already imports but keep file cohesive
    repl := strings.NewReplacer(" ", "+")
    return repl.Replace(s)
}
