package main

import (
    "archive/zip"
    "io"
    "os"
    "path/filepath"
    "strings"
)

// saveToFile writes the reader to the given file path, creating directories as needed.
func saveToFile(r io.Reader, path string) error {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && !os.IsExist(err) {
        // ignore error if directory already exists
    }
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()
    _, err = io.Copy(f, r)
    return err
}

// unzip extracts a zip file to destDir, preserving modes and structure.
func unzip(zipPath, destDir string) error {
    zr, err := zip.OpenReader(zipPath)
    if err != nil {
        return err
    }
    defer zr.Close()

    if err := os.MkdirAll(destDir, 0o755); err != nil {
        return err
    }

    // Detect if the zip contains a single top-level directory that matches
    // the destination base directory name. If so, strip that component to
    // avoid nested same-name directories like destDir/destDir/...
    base := filepath.Base(destDir)
    topLevels := make(map[string]struct{})
    for _, f := range zr.File {
        // Normalize separators in zip entries
        name := strings.TrimLeft(strings.ReplaceAll(f.Name, "\\", "/"), "/")
        if name == "" {
            continue
        }
        // Extract first path component
        if i := strings.IndexByte(name, '/'); i >= 0 {
            topLevels[name[:i]] = struct{}{}
        } else {
            topLevels[name] = struct{}{}
        }
    }

    var stripPrefix string
    if len(topLevels) == 1 {
        for tl := range topLevels {
            if tl == base {
                stripPrefix = tl + "/"
            }
        }
    }

    for _, f := range zr.File {
        // Normalize and optionally strip the top-level prefix
        name := strings.TrimLeft(strings.ReplaceAll(f.Name, "\\", "/"), "/")
        if stripPrefix != "" {
            name = strings.TrimPrefix(name, stripPrefix)
        }
        if name == "" {
            // nothing to create (e.g., top-level dir entry when stripped)
            continue
        }
        p := filepath.Join(destDir, name)
        if f.FileInfo().IsDir() {
            if err := os.MkdirAll(p, f.Mode()); err != nil {
                return err
            }
            continue
        }
        if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
            return err
        }
        rc, err := f.Open()
        if err != nil {
            return err
        }
        w, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
        if err != nil {
            rc.Close()
            return err
        }
        if _, err := io.Copy(w, rc); err != nil {
            rc.Close()
            w.Close()
            return err
        }
        rc.Close()
        w.Close()
    }
    return nil
}
