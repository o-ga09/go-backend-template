package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if len(cfg.Paths) == 0 {
		t.Fatal("loadConfig() Paths が空")
	}
	if len(cfg.Replacements) == 0 {
		t.Fatal("loadConfig() Replacements が空")
	}
	for _, r := range cfg.Replacements {
		if _, err := embeddedTemplates.ReadFile("templates/" + r.Template); err != nil {
			t.Errorf("replacement %q のテンプレート %q を読み込めない: %v", r.Target, r.Template, err)
		}
	}
}

func TestParseYesNo(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"小文字yは許可", "y", true},
		{"yesも許可", "yes", true},
		{"大文字YESも許可", "YES", true},
		{"前後の空白は無視する", "  y  ", true},
		{"空文字はNoとして扱う", "", false},
		{"nはNoとして扱う", "n", false},
		{"それ以外の文字列はNoとして扱う", "maybe", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseYesNo(tc.input); got != tc.want {
				t.Errorf("parseYesNo(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestFilterExisting(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal/domain/product"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal/domain/product/entity.go"), []byte("package product"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := filterExisting(root, []string{
		"internal/domain/product",
		"internal/domain/cart", // 存在しない
	})

	if len(got) != 1 || got[0] != "internal/domain/product" {
		t.Errorf("filterExisting() = %v, want [internal/domain/product]", got)
	}
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"yで承認", "y\n", true},
		{"改行なしの入力もEOFで承認しない", "", false},
		{"noで拒否", "n\n", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := strings.NewReader(tc.input)
			var out bytes.Buffer
			got, err := confirm(in, &out)
			if err != nil {
				t.Fatalf("confirm() error = %v", err)
			}
			if got != tc.want {
				t.Errorf("confirm() = %v, want %v", got, tc.want)
			}
			if !strings.Contains(out.String(), "よろしいですか") {
				t.Errorf("confirm() のプロンプトが出力されていない: %q", out.String())
			}
		})
	}
}

// setupGitRepo は destroy-sample が git rm / git add を実行できる最小限の
// gitリポジトリを一時ディレクトリに作る。
func setupGitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "test")
	return root
}

func writeAndCommit(t *testing.T, root, relPath, content string) {
	t.Helper()
	full := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("add", "--", relPath)
	run("commit", "-q", "-m", "add "+relPath)
}

func TestRun_DryRunDoesNotModifyRepo(t *testing.T) {
	root := setupGitRepo(t)
	writeAndCommit(t, root, "internal/domain/product/entity.go", "package product")
	writeAndCommit(t, root, "internal/router/route.go", "package router // original")

	cfg := config{
		Paths: []string{"internal/domain/product"},
		Replacements: []replacement{
			{Target: "internal/router/route.go", Template: "router_route.go.tmpl"},
		},
	}

	var out bytes.Buffer
	if err := run(root, cfg, true, true, strings.NewReader(""), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "internal/domain/product/entity.go")); err != nil {
		t.Errorf("dry-runなのにファイルが削除されている: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "internal/router/route.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "package router // original" {
		t.Errorf("dry-runなのにファイルが上書きされている: %q", got)
	}
	if !strings.Contains(out.String(), "internal/domain/product") {
		t.Errorf("dry-run出力に削除対象が含まれていない: %q", out.String())
	}
}

func TestRun_DeletesAndReplaces(t *testing.T) {
	root := setupGitRepo(t)
	writeAndCommit(t, root, "internal/domain/product/entity.go", "package product")
	writeAndCommit(t, root, "internal/router/route.go", "package router // original")

	cfg := config{
		Paths: []string{"internal/domain/product", "internal/domain/cart"}, // cartは存在しない
		Replacements: []replacement{
			{Target: "internal/router/route.go", Template: "router_route.go.tmpl"},
		},
	}

	var out bytes.Buffer
	if err := run(root, cfg, false, true, strings.NewReader(""), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "internal/domain/product")); !os.IsNotExist(err) {
		t.Errorf("削除対象が残っている: err = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, "internal/router/route.go"))
	if err != nil {
		t.Fatal(err)
	}
	wantPrefix := "package router\n"
	if !strings.HasPrefix(string(got), wantPrefix) {
		t.Errorf("テンプレートで上書きされていない: %q", got)
	}
}

func TestRun_AbortsWhenConfirmationDeclined(t *testing.T) {
	root := setupGitRepo(t)
	writeAndCommit(t, root, "internal/domain/product/entity.go", "package product")

	cfg := config{Paths: []string{"internal/domain/product"}}

	var out bytes.Buffer
	if err := run(root, cfg, false, false, strings.NewReader("n\n"), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "internal/domain/product/entity.go")); err != nil {
		t.Errorf("確認を拒否したのにファイルが削除されている: %v", err)
	}
}
