package main

import (
    "net/url"
    "reflect"
    "testing"
)

func TestSanitizePackage(t *testing.T) {
    cases := []struct{ in, out string }{
        {"Com.Example.Demo", "com.example.demo"},
        {"com-example demo", "comexampledemo"},
        {"..com..example..demo..", "com.example.demo"},
        {"-A-.-B-", "a.b"},
        {"", ""},
    }
    for _, c := range cases {
        if got := sanitizePackage(c.in); got != c.out {
            t.Errorf("sanitizePackage(%q) = %q; want %q", c.in, got, c.out)
        }
    }
}

func TestBuildURL(t *testing.T) {
    o := options{
        baseURL:     "https://start.spring.io",
        target:      "zip",
        projectType: "maven-project",
        language:    "java",
        bootVersion: "3.3.4",
        groupID:     "com.example",
        artifactID:  "demo",
        name:        "demo",
        description: "Demo project",
        packageName: "com.example.demo",
        packaging:   "jar",
        javaVersion: "21",
        dependencies:"web,data-jpa , security",
        baseDir:     "demo",
    }
    u, err := buildURL(o)
    if err != nil {
        t.Fatalf("buildURL error: %v", err)
    }
    parsed, err := url.Parse(u)
    if err != nil {
        t.Fatalf("url.Parse error: %v", err)
    }
    if parsed.Scheme != "https" || parsed.Host != "start.spring.io" || parsed.Path != "/starter.zip" {
        t.Fatalf("unexpected URL: %s", u)
    }
    got := parsed.Query()
    want := url.Values{
        "type":        []string{"maven-project"},
        "language":    []string{"java"},
        "bootVersion": []string{"3.3.4"},
        "baseDir":     []string{"demo"},
        "groupId":     []string{"com.example"},
        "artifactId":  []string{"demo"},
        "name":        []string{"demo"},
        "description": []string{"Demo project"},
        "packageName": []string{"com.example.demo"},
        "packaging":   []string{"jar"},
        "javaVersion": []string{"21"},
        "dependencies": []string{"web,data-jpa,security"},
    }
    if !reflect.DeepEqual(got, want) {
        t.Fatalf("query mismatch:\n got: %#v\nwant: %#v", got, want)
    }
}

