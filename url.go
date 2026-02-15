package main

import (
	"errors"
	"net/url"
	"regexp"
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
	// Normalize historical suffix styles (e.g., .RELEASE, .BUILD-SNAPSHOT, .M7, .RC1)
	add("bootVersion", normalizeBootVersion(o.bootVersion))
	add("baseDir", o.baseDir)
	add("groupId", o.groupID)
	add("artifactId", o.artifactID)
	add("name", o.name)
	add("description", o.description)
	add("packageName", o.packageName)
	add("packaging", o.packaging)
	add("javaVersion", o.javaVersion)
	add("configurationFileFormat", o.configFileFormat)

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

// normalizeBootVersion converts historical Spring Boot version notations
// to the forms accepted by modern Initializr servers.
// Examples:
//   - 3.5.5.RELEASE        -> 3.5.5
//   - 2.0.0.BUILD-SNAPSHOT -> 2.0.0-SNAPSHOT
//   - 2.0.0.M7             -> 2.0.0-M7
//   - 2.0.0.RC1            -> 2.0.0-RC1
func normalizeBootVersion(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	// RELEASE suffix: drop it (accept both ".RELEASE" and "-RELEASE")
	s = strings.TrimSuffix(s, ".RELEASE")
	s = strings.TrimSuffix(s, "-RELEASE")

	// BUILD-SNAPSHOT: convert to -SNAPSHOT
	s = strings.ReplaceAll(s, ".BUILD-SNAPSHOT", "-SNAPSHOT")
	s = strings.ReplaceAll(s, "-BUILD-SNAPSHOT", "-SNAPSHOT")
	// Rare: handle ".SNAPSHOT" -> "-SNAPSHOT"
	if strings.HasSuffix(s, ".SNAPSHOT") {
		s = strings.TrimSuffix(s, ".SNAPSHOT") + "-SNAPSHOT"
	}

	// Convert ".M<digits>" and ".RC<digits>" at the end to hyphenated form.
	re := regexp.MustCompile(`\.(M|RC)(\d+)$`)
	s = re.ReplaceAllString(s, "-$1$2")

	return s
}
