package main

import (
    "errors"
    "net/url"
    "strings"
)

// buildURL constructs the Initializr starter URL from options.
func buildURL(o options) (string, error) {
    if o.baseURL == "" {
        return "", errors.New("base-url must not be empty")
    }
    base := strings.TrimRight(o.baseURL, "/") + "/starter."
    switch strings.ToLower(o.target) {
    case "zip":
        base += "zip"
    default:
        return "", errors.New("unsupported target: " + o.target)
    }

    q := url.Values{}
    add := func(k, v string) {
        if strings.TrimSpace(v) != "" {
            q.Set(k, v)
        }
    }

    add("type", o.projectType)
    add("language", o.language)
    add("bootVersion", o.bootVersion)
    add("baseDir", o.baseDir)
    add("groupId", o.groupID)
    add("artifactId", o.artifactID)
    add("name", o.name)
    add("description", o.description)
    add("packageName", o.packageName)
    add("packaging", o.packaging)
    add("javaVersion", o.javaVersion)

    deps := strings.TrimSpace(o.dependencies)
    if deps != "" {
        // normalize: remove whitespace around commas
        parts := strings.Split(deps, ",")
        for i := range parts {
            parts[i] = strings.TrimSpace(parts[i])
        }
        q.Set("dependencies", strings.Join(parts, ","))
    }

    return base + "?" + q.Encode(), nil
}

// sanitizePackage normalizes a Java package name string from user inputs.
func sanitizePackage(s string) string {
    s = strings.ToLower(s)
    replacer := strings.NewReplacer("-", "", " ", "")
    s = replacer.Replace(s)
    for strings.Contains(s, "..") {
        s = strings.ReplaceAll(s, "..", ".")
    }
    s = strings.Trim(s, ".")
    return s
}

