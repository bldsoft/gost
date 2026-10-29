package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunWritesRenamedService(t *testing.T) {
	dir := t.TempDir()
	err := run(options{
		module: "example.com/acme/widgets",
		name:   "widgets",
		dir:    dir,
		noGo:   true,
	})
	require.NoError(t, err)

	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		body, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		require.NotContains(t, string(body), "bldsoft/gost/examples", path)

		return nil
	})
	require.NoError(t, err)

	mainGo, err := os.ReadFile(filepath.Join(dir, "cmd", "widgets", "main.go"))
	require.NoError(t, err)
	require.Contains(t, string(mainGo), "example.com/acme/widgets/pkg/config")
	_, err = os.Stat(filepath.Join(dir, "cmd", "notes"))
	require.True(t, os.IsNotExist(err))

	cfg, err := os.ReadFile(filepath.Join(dir, "pkg", "config", "config.go"))
	require.NoError(t, err)
	require.Contains(t, string(cfg), `const ServiceName = "widgets"`)

	mockery, err := os.ReadFile(filepath.Join(dir, ".mockery.yml"))
	require.NoError(t, err)
	require.Contains(t, string(mockery), "example.com/acme/widgets/pkg/service:")

	lint, err := os.ReadFile(filepath.Join(dir, ".golangci.yml"))
	require.NoError(t, err)
	require.Contains(t, string(lint), "prefix(example.com/acme)")
	require.NotContains(t, string(lint), "prefix(github.com/bldsoft)")
}

func TestRunRejectsNonEmptyDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "keep"), []byte("x"), 0o644))

	err := run(options{module: "example.com/acme/widgets", name: "widgets", dir: dir, noGo: true})
	require.ErrorContains(t, err, "not empty")
}

func TestRunRejectsName(t *testing.T) {
	err := run(options{module: "example.com/acme/widgets", name: "Widgets", dir: t.TempDir(), noGo: true})
	require.ErrorContains(t, err, "invalid -name")
}
