package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func parseFlags() options {
	var o options
	// If invoked without any arguments, default to interactive mode.
	// This is applied after flag.Parse so explicit flags still override.
	noArgs := len(os.Args) == 1

	flag.StringVar(&o.baseURL, "base-url", defaultBaseURL, "Spring Initializr base URL")
	flag.StringVar(&o.target, "target", "zip", "Archive format: zip (default)")
	flag.StringVar(&o.projectType, "type", "maven-project", "Project type: maven-project, gradle-project, or gradle-build")
	flag.StringVar(&o.language, "language", "java", "Language: java, kotlin, or groovy")
	flag.StringVar(&o.bootVersion, "boot-version", "", "Spring Boot version (optional)")
	flag.StringVar(&o.groupID, "group-id", "com.example", "Group ID")
	flag.StringVar(&o.artifactID, "artifact-id", "demo", "Artifact ID")
	flag.StringVar(&o.name, "name", "demo", "Project name")
	flag.StringVar(&o.description, "description", "Demo project for Spring Boot", "Project description")
	flag.StringVar(&o.packageName, "package-name", "", "Base package name (default: groupId + '.' + artifactId)")
	flag.StringVar(&o.packaging, "packaging", "jar", "Packaging: jar or war")
	flag.StringVar(&o.javaVersion, "java-version", "", "Java version (optional). If omitted, server default is used")
	flag.StringVar(&o.configFileFormat, "configuration-file-format", "", "Configuration file format: properties or yaml (optional)")
	flag.StringVar(&o.dependencies, "dependencies", "", "Comma-separated dependency IDs, e.g. web,data-jpa,postgresql")
	flag.StringVar(&o.baseDir, "base-dir", "", "Project root directory name (default: artifactId)")

	flag.StringVar(&o.output, "output", "", "Output zip file path (default: <artifactId>.zip)")
	flag.BoolVar(&o.extract, "extract", false, "Extract archive into directory (uses base-dir)")
	flag.BoolVar(&o.dryRun, "dry-run", false, "Print the generated URL and exit")
	flag.IntVar(&o.timeout, "timeout", 60, "Download timeout in seconds")
	flag.BoolVar(&o.verbose, "v", false, "Verbose output")
	flag.BoolVar(&o.interactive, "interactive", false, "Interactive TUI mode")
	flag.BoolVar(&o.interactive, "i", false, "Interactive TUI mode (shorthand)")
	flag.BoolVar(&o.showVersion, "version", false, "Print version and exit")
	flag.BoolVar(&o.showVersion, "V", false, "Print version and exit (shorthand)")
	flag.BoolVar(&o.showLicense, "license", false, "Print licenses (app + NOTICE) and exit")
	flag.BoolVar(&o.showLicense, "L", false, "Print licenses (app + NOTICE) and exit (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Spring Initializr CLI (Go)\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s --type maven-project --language java \\\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "     --group-id com.example --artifact-id demo \\\n")
		fmt.Fprintf(os.Stderr, "     --dependencies web,data-jpa --extract\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nNotes:\n- Dependencies are Spring Initializr IDs (e.g. web, data-jpa, security).\n")
		fmt.Fprintf(os.Stderr, "- If --extract is set, the zip will be downloaded and extracted into --base-dir (defaults to artifact-id).\n")
		fmt.Fprintf(os.Stderr, "- Use --dry-run to just print the URL.\n")
		fmt.Fprintf(os.Stderr, "- Use --version or -V to print the version.\n")
		fmt.Fprintf(os.Stderr, "- Use --license or -L to print licenses and exit.\n")
	}

	flag.Parse()

	if noArgs {
		o.interactive = true
	}

	// Fill derived defaults
	if o.baseDir == "" {
		o.baseDir = o.artifactID
	}
	if o.packageName == "" {
		o.packageName = sanitizePackage(o.groupID + "." + o.artifactID)
	} else {
		o.packageName = sanitizePackage(o.packageName)
	}
	if o.output == "" {
		o.output = o.artifactID + ".zip"
	}

	// Normalize some shortcuts
	switch strings.ToLower(o.projectType) {
	case "maven", "maven-project":
		o.projectType = "maven-project"
	case "gradle", "gradle-project":
		o.projectType = "gradle-project"
	case "gradle-build":
		// as-is
	default:
		// keep as provided, server will validate
	}

	return o
}
