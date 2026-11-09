package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/bldsoft/gost/examples"
)

const (
	notesImport = "github.com/bldsoft/gost/examples/notes"
	gostModule  = "github.com/bldsoft/gost"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

type options struct {
	module  string
	name    string
	dir     string
	gost    string
	replace string
	noGo    bool
}

func main() {
	var opts options
	flag.StringVar(&opts.module, "module", "", "module path of the new service, for example github.com/acme/widgets")
	flag.StringVar(&opts.name, "name", "", "service name: binary, URL prefix and database name (default: the last element of -module)")
	flag.StringVar(&opts.dir, "dir", "", "output directory (default: ./<name>)")
	flag.StringVar(&opts.gost, "gost", "", "gost module version to require (default: the version of this command, or latest)")
	flag.StringVar(&opts.replace, "replace", "", "local gost checkout used as a replace directive")
	flag.BoolVar(&opts.noGo, "no-go", false, "write files only, and skip go mod, go generate and config.md")
	flag.Parse()

	if err := run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "scaffold: %v\n", err)
		os.Exit(1)
	}
}

func run(opts options) error {
	if err := opts.normalize(); err != nil {
		return err
	}
	if err := emptyDir(opts.dir); err != nil {
		return err
	}
	if err := writeNotes(opts); err != nil {
		return err
	}
	if err := writeScaffold(opts); err != nil {
		return err
	}
	if opts.noGo {
		return nil
	}

	return setupModule(opts)
}

func (o *options) normalize() error {
	if o.module == "" {
		return errors.New("-module is required")
	}
	if path.Dir(o.module) == "." || strings.Contains(o.module, " ") || strings.Contains(o.module, "..") {
		return fmt.Errorf("invalid module path %q", o.module)
	}
	if o.name == "" {
		o.name = path.Base(o.module)
	}
	if !namePattern.MatchString(o.name) {
		return fmt.Errorf("invalid -name %q: must match %s", o.name, namePattern)
	}
	if o.dir == "" {
		o.dir = o.name
	}
	if o.gost == "" {
		o.gost = gostVersion()
	}

	return nil
}

func gostVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || !strings.HasPrefix(info.Main.Version, "v") {
		return "latest"
	}

	return info.Main.Version
}

func emptyDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("%s is not empty", dir)
	}

	return nil
}

func writeNotes(opts options) error {
	return fs.WalkDir(examples.Files, "notes", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(p, "notes/")
		if rel == "notes" || skipNotes(rel) {
			if d.IsDir() && (rel == "bin" || strings.HasPrefix(rel, "bin/")) {
				return fs.SkipDir
			}

			return nil
		}
		if rel == "cmd/notes" || strings.HasPrefix(rel, "cmd/notes/") {
			rel = "cmd/" + opts.name + strings.TrimPrefix(rel, "cmd/notes")
		}
		dest := filepath.Join(opts.dir, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		data, err := examples.Files.ReadFile(p)
		if err != nil {
			return err
		}
		text := strings.ReplaceAll(string(data), notesImport, opts.module)
		if rel == "pkg/config/config.go" {
			text = strings.Replace(text, `const ServiceName = "notes"`, `const ServiceName = "`+opts.name+`"`, 1)
		}

		return writeFile(dest, []byte(text))
	})
}

func skipNotes(rel string) bool {
	switch rel {
	case "README.md", "Makefile", "Dockerfile", "bin":
		return true
	default:
		return strings.HasPrefix(rel, "bin/")
	}
}

func writeScaffold(opts options) error {
	owner := path.Dir(opts.module)
	org := path.Base(owner)

	return fs.WalkDir(examples.Files, "scaffold", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := strings.TrimPrefix(p, "scaffold/")
		data, err := examples.Files.ReadFile(p)
		if err != nil {
			return err
		}
		text := strings.ReplaceAll(string(data), "<module>", opts.module)
		text = strings.ReplaceAll(text, "<org>", org)
		text = strings.ReplaceAll(text, "<svc>", opts.name)
		if rel == "golangci.yml" {
			text = strings.ReplaceAll(text, "prefix(github.com/bldsoft)", "prefix("+owner+")")
		}

		return writeFile(filepath.Join(opts.dir, scaffoldDest(rel)), []byte(text))
	})
}

func scaffoldDest(rel string) string {
	switch rel {
	case "gitignore":
		return ".gitignore"
	case "golangci.yml":
		return ".golangci.yml"
	case "mockery.yml":
		return ".mockery.yml"
	case "github-workflow.yml":
		return filepath.Join(".github", "workflows", "ci.yml")
	default:
		return rel
	}
}

func writeFile(dest string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	return os.WriteFile(dest, data, 0o644)
}

func setupModule(opts options) error {
	steps := [][]string{{"go", "mod", "init", opts.module}}
	if opts.replace != "" {
		abs, err := filepath.Abs(opts.replace)
		if err != nil {
			return err
		}
		steps = append(steps, []string{"go", "mod", "edit", "-replace", gostModule + "=" + abs})
	}
	steps = append(steps,
		[]string{"go", "get", gostModule + "@" + opts.gost},
		[]string{"go", "get", "github.com/hashicorp/go-metrics@v0.5.4"},
		[]string{"go", "get", "-tool", "github.com/vektra/mockery/v3@latest"},
		[]string{"go", "get", "-tool", "github.com/dmarkham/enumer@latest"},
		[]string{"go", "generate", "./..."},
		[]string{"go", "mod", "tidy"},
		[]string{"go", "test", "-count=1", "./pkg/config", "-run", "TestConfigDescriptionGolden"},
		[]string{"golangci-lint", "fmt", "./..."},
	)
	for _, step := range steps {
		env := os.Environ()
		if step[0] == "go" && step[1] == "test" {
			env = append(env, "UPDATE_CONFIG_MD=1")
		}
		if err := runCmd(opts.dir, env, step...); err != nil {
			return err
		}
	}

	return nil
}

func runCmd(dir string, env []string, args ...string) error {
	fmt.Fprintf(os.Stderr, "+ %s\n", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", strings.Join(args, " "), err)
	}

	return nil
}
