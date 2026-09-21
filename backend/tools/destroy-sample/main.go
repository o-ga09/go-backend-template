// Command destroy-sample removes the EC sample implementation shipped with
// this template (product/cart/order domain, handlers, migrations, seed data)
// and restores the router to an empty skeleton, so the repository can be
// reused as the base for a new project.
//
// It must be run inside a git-managed, uncommitted clone: deletions go
// through `git rm` (never `rm -rf`) so the result can always be inspected
// with `git diff` / `git status` and reverted with `git checkout`.
package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed config.yml
var embeddedConfig embed.FS

//go:embed templates
var embeddedTemplates embed.FS

type replacement struct {
	Target   string `yaml:"target"`
	Template string `yaml:"template"`
}

type config struct {
	Paths        []string      `yaml:"paths"`
	Replacements []replacement `yaml:"replacements"`
}

func main() {
	dryRun := flag.Bool("dry-run", false, "削除対象を表示するだけで、実際の削除・書き換えは行わない")
	yes := flag.Bool("yes", false, "確認プロンプトをスキップする")
	force := flag.Bool("force", false, "--yes の別名")
	flag.Parse()

	cfg, err := loadConfig()
	if err != nil {
		fail(err)
	}

	repoRoot, err := gitRepoRoot()
	if err != nil {
		fail(fmt.Errorf("Gitリポジトリのルートを特定できませんでした。Git管理下のクローンで実行してください: %w", err))
	}

	if err := run(repoRoot, cfg, *dryRun, *yes || *force, os.Stdin, os.Stdout); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "destroy-sample: "+err.Error())
	os.Exit(1)
}

func loadConfig() (config, error) {
	b, err := embeddedConfig.ReadFile("config.yml")
	if err != nil {
		return config{}, fmt.Errorf("config.ymlの読み込みに失敗しました: %w", err)
	}
	var cfg config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return config{}, fmt.Errorf("config.ymlのパースに失敗しました: %w", err)
	}
	return cfg, nil
}

func gitRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func run(repoRoot string, cfg config, dryRun, skipConfirm bool, in io.Reader, out io.Writer) error {
	existing := filterExisting(repoRoot, cfg.Paths)

	fmt.Fprintln(out, "=== 削除対象パス ===")
	if len(existing) == 0 {
		fmt.Fprintln(out, "(該当するパスはありません。既に実行済みか、config.ymlを確認してください)")
	}
	for _, p := range existing {
		fmt.Fprintln(out, "  - "+p)
	}

	fmt.Fprintln(out, "\n=== スケルトンで上書きするファイル ===")
	for _, r := range cfg.Replacements {
		fmt.Fprintln(out, "  - "+r.Target)
	}

	if dryRun {
		fmt.Fprintln(out, "\ndry-run: 実際の削除・書き換えは行いませんでした。")
		return nil
	}

	if len(existing) == 0 && len(cfg.Replacements) == 0 {
		return nil
	}

	if !skipConfirm {
		ok, err := confirm(in, out)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(out, "中断しました。")
			return nil
		}
	}

	if len(existing) > 0 {
		args := append([]string{"-C", repoRoot, "rm", "-r", "--ignore-unmatch", "--"}, existing...)
		if o, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			return fmt.Errorf("git rmに失敗しました: %w\n%s", err, o)
		}
	}

	for _, r := range cfg.Replacements {
		if err := applyReplacement(repoRoot, r); err != nil {
			return err
		}
	}

	fmt.Fprintln(out, "\n削除・書き換えが完了しました。`git status` / `git diff --cached` で内容を確認してください。")
	fmt.Fprintln(out, "続けて `cd backend && go build ./...` と `cd frontend && pnpm build` が壊れていないことを確認してください。")
	return nil
}

// filterExisting はリポジトリ上に実在するパスだけを返す。config.ymlは他のIssue
// (フロントエンドEC画面等)の実装と並行して更新されるため、未実装分のパスが
// 含まれていても構わない設計にしている。
func filterExisting(repoRoot string, paths []string) []string {
	var existing []string
	for _, p := range paths {
		if _, err := os.Lstat(filepath.Join(repoRoot, p)); err == nil {
			existing = append(existing, p)
		}
	}
	return existing
}

func confirm(in io.Reader, out io.Writer) (bool, error) {
	fmt.Fprint(out, "\n上記を削除・書き換えします。よろしいですか? [y/N]: ")
	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		return false, scanner.Err()
	}
	return parseYesNo(scanner.Text()), nil
}

func parseYesNo(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

func applyReplacement(repoRoot string, r replacement) error {
	content, err := embeddedTemplates.ReadFile(path.Join("templates", r.Template))
	if err != nil {
		return fmt.Errorf("テンプレート%sの読み込みに失敗しました: %w", r.Template, err)
	}
	target := filepath.Join(repoRoot, r.Target)
	if err := os.WriteFile(target, content, 0o644); err != nil {
		return fmt.Errorf("%sへの書き込みに失敗しました: %w", r.Target, err)
	}
	if o, err := exec.Command("git", "-C", repoRoot, "add", "--", r.Target).CombinedOutput(); err != nil {
		return fmt.Errorf("git addに失敗しました: %w\n%s", err, o)
	}
	return nil
}
