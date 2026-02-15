package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultBaseURL = "https://start.spring.io"

// version is set at build time via -ldflags "-X main.version=<version>"
var version = "dev"

type options struct {
	baseURL          string
	target           string // zip or tgz (only zip implemented for now)
	projectType      string // maven-project, gradle-project, gradle-build
	language         string // java, kotlin, groovy
	bootVersion      string
	groupID          string
	artifactID       string
	name             string
	description      string
	packageName      string
	packaging        string // jar or war
	javaVersion      string // 8, 11, 17, 21, etc.
	configFileFormat string // properties or yaml
	dependencies     string // comma-separated
	baseDir          string

	output  string // output file path for zip
	extract bool   // extract zip to directory (baseDir)
	dryRun  bool   // print URL and exit
	timeout int    // seconds
	verbose bool

	// interactive control (not a flag)
	interactive bool

	// show version and exit
	showVersion bool

	// show license and notices and exit
	showLicense bool
}

func main() {
	opts := parseFlags()
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

/*
	 parseFlags moved to parser.go
		var o options

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
		flag.StringVar(&o.javaVersion, "java-version", "21", "Java version, e.g. 17 or 21")
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
*/
func run(o options) error {
	if o.showVersion {
		fmt.Println(version)
		return nil
	}
	if o.showLicense {
		printLicenses()
		return nil
	}
	if o.interactive {
		// Use the full-featured TUI if available
		return runInteractive(o)
	}
	if strings.ToLower(o.target) != "zip" {
		return fmt.Errorf("unsupported target '%s' (only 'zip' is supported)", o.target)
	}

	u, err := buildURL(o)
	if err != nil {
		return err
	}

	if o.dryRun {
		fmt.Println(u)
		return nil
	}

	if o.verbose {
		fmt.Println("Downloading:", u)
	}

	client := &http.Client{Timeout: time.Duration(o.timeout) * time.Second}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/zip, application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("bad response: %s\n%s", resp.Status, string(b))
	}

	if o.extract {
		// Download to temp file, then unzip into baseDir
		tmpf, err := os.CreateTemp("", "spring-initializr-*.zip")
		if err != nil {
			return err
		}
		tmp := tmpf.Name()
		if o.verbose {
			fmt.Println("Saving to temp:", tmp)
		}
		if _, err := io.Copy(tmpf, resp.Body); err != nil {
			tmpf.Close()
			os.Remove(tmp)
			return err
		}
		tmpf.Close()

		if err := unzip(tmp, o.baseDir); err != nil {
			os.Remove(tmp)
			return err
		}
		os.Remove(tmp)
		if o.verbose {
			fmt.Println("Extracted into:", o.baseDir)
		}
		return nil
	}

	// Save zip to file
	if err := saveToFile(resp.Body, o.output); err != nil {
		return err
	}
	if o.verbose {
		fmt.Println("Saved:", o.output)
	}
	return nil
}

// runInteractive provides a simple full-screen, line-based TUI without external deps.
func applyAction(o options, action string) options {
	switch action {
	case "download":
		o.dryRun = false
		// extract as chosen
	case "extract":
		o.dryRun = false
		o.extract = true
	}
	o.interactive = false
	return o
}

// URL construction and helpers moved to url.go; filesystem helpers moved to fs.go
