package cmd

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestGetKubeconfigsDir(t *testing.T) {
	home := "/home/johndoe"
	expected := fmt.Sprintf("%s/.kube/configs", home)
	os.Setenv("HOME", home)
	got, err := getKubeconfigsDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != expected {
		t.Fatalf("expected=%q; got=%q", expected, got)
	}
}

func TestSetOptsFromEnv(t *testing.T) {
	dir := "/home/johndoe/.kube/configs"
	height := "75%"
	enabled := "1"

	os.Setenv("KCNF_DIR", dir)
	os.Setenv("KCNF_DIR_HEIGHT", height)
	os.Setenv("KCNF_NO_VERBOSE", enabled)
	os.Setenv("KCNF_NO_SHELL", enabled)
	os.Setenv("KCNF_COPY_CLIP", enabled)
	os.Setenv("KCNF_SYMLINK", enabled)

	if err := setOptsFromEnv(); err != nil {
		t.Fatal(err)
	}

	if opts.kubeconfigsDir != dir {
		t.Fatalf("opts.kubeconfigsDir expected=%q; got=%q", dir, opts.kubeconfigsDir)
	} else if opts.selectionHeight != height {
		t.Fatalf("opts.selectionHeight expected=%q; got=%q", height, opts.selectionHeight)
	} else if !opts.quietFlag {
		t.Fatalf("opts.noVerboseFlag expected=%t; got=%t", true, opts.quietFlag)
	} else if !opts.printFlag {
		t.Fatalf("opts.noShellFlag expected=%t; got=%t", true, opts.printFlag)
	} else if !opts.clipboardFlag {
		t.Fatalf("opts.copyClipFlag expected=%t; got=%t", true, opts.clipboardFlag)
	} else if !opts.symlinkFlag {
		t.Fatalf("opts.symlinkFlag expected=%t; got=%t", true, opts.symlinkFlag)
	}
}

func TestGetCurrentContext(t *testing.T) {
	expected := "testing-eu-01"
	body := []byte(fmt.Sprintf("current-context: %s\n", expected))

	file, err := os.CreateTemp(os.TempDir(), "kcnf_kubeconfig_")
	if err != nil {
		t.Fatalf("failed to create tmp file: %v", err)
	}
	path := file.Name()
	defer os.Remove(path)

	if err := os.WriteFile(path, body, 0644); err != nil {
		t.Fatalf("failed to write to tmp file: %v", err)
	}

	if got := getCurrentContext(path); got != expected {
		t.Fatalf("expected=%q; got=%q", expected, got)
	}
}

func TestGetKubeconfigs(t *testing.T) {
	ctx := "testing-eu-01"
	body := []byte(fmt.Sprintf("current-context: %s\n", ctx))

	dir, err := os.MkdirTemp(os.TempDir(), "kcnf_configs_")
	if err != nil {
		t.Fatalf("failed to create tmp directory: %v", err)
	}
	defer os.Remove(dir)

	file, err := os.CreateTemp(dir, "kcnf_kubeconfig_")
	if err != nil {
		t.Fatalf("failed to create tmp file: %v", err)
	}
	path := file.Name()
	defer os.Remove(path)

	if err := os.WriteFile(path, body, 0644); err != nil {
		t.Fatalf("failed to write to tmp file: %v", err)
	}

	got, err := getKubeconfigs(dir)
	if err != nil {
		t.Fatal(err)
	}

	expected := [][]string{{ctx, path}}
	if got == nil {
		t.Fatalf("expected=%q; got=%q", expected, got)
	}
}

func TestGetPreviewCmd(t *testing.T) {
	expected := "{ echo '# {2}'; kubectl config view --kubeconfig {2}; } | bat --style=plain --color=always --language=yaml"
	if got := getPreviewCmd(); got != expected {
		t.Fatalf("expected=%q; got=%q", expected, got)
	}
}

func TestConfigureFzf(t *testing.T) {
	height := "40%"
	query := "test"
	preview := "{ echo '# {2}'; kubectl config view --kubeconfig {2}; } | cat"
	if _, err := configureFzf(height, query, preview); err != nil {
		t.Fatal(err)
	}
}

func TestSymlinkKubeconfig(t *testing.T) {
	os.Setenv("HOME", os.TempDir())
	dir := fmt.Sprintf("%s/.kube", os.TempDir())
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatalf("failed to create tmp directory: %v", err)
	}
	defer os.RemoveAll(dir)
	file, err := os.CreateTemp(os.TempDir(), "testing-eu-01")
	if err != nil {
		t.Fatalf("failed to create tmp file: %v", err)
	}
	kubeconfig := file.Name()
	if err := symlinkKubeconfig(kubeconfig); err != nil {
		t.Fatal(err)
	}
}

func TestLaunchShell(t *testing.T) {
	shell := os.Getenv("SHELL")
	kubeconfig := "/home/johndoe/.kube/configs/testing-eu-01"
	kubecontext := "testing-eu-01"
	if err := launchShell(shell, kubeconfig, kubecontext); err != nil {
		t.Fatal(err)
	}
}

func TestProcessSelection(t *testing.T) {
	selecion := "testing-eu-01\t/home/johndoe/.kube/configs/testing-eu-01"
	opts.printFlag = true
	opts.quietFlag = false
	opts.clipboardFlag = false
	opts.symlinkFlag = false

	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()
	r, w, _ := os.Pipe()
	os.Stdout = w

	if err := processSelection(selecion); err != nil {
		t.Fatal(err)
	}

	w.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed reading stdout: %v", err)
	}

	expected := "⮺ testing-eu-01\nexport KUBECONFIG='/home/johndoe/.kube/configs/testing-eu-01'\n"
	if got := buf.String(); got != expected {
		t.Errorf("expected=%q; got=%q", expected, got)
	}
}
