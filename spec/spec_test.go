// Copyright 2026 Ibrahim Karakira
// SPDX-License-Identifier: Apache-2.0

// Package spec holds the predicate specifications. These tests keep the
// Markdown honest: the example must be a full Statement of our type, every
// field in it must be documented, and relative links must resolve.
package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const (
	specFile = "agent-work-v0.md"
	typeURI  = "https://ibrahimkarakira.github.io/agent-signoff/predicate/agent-work/v0"
)

func readSpec(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(specFile)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	return string(b)
}

// exampleStatement returns the first ```json block in the spec, decoded.
func exampleStatement(t *testing.T, doc string) map[string]any {
	t.Helper()
	m := regexp.MustCompile("(?s)```json\r?\n(.*?)```").FindStringSubmatch(doc)
	if m == nil {
		t.Fatal("spec has no ```json example")
	}
	var st map[string]any
	if err := json.Unmarshal([]byte(m[1]), &st); err != nil {
		t.Fatalf("example is not valid JSON: %v", err)
	}
	return st
}

// fieldPaths flattens an object into dotted paths, arrays as [*].
// DigestSet maps (any key named "digest" or ending in "Digest") are leaves:
// their keys are algorithm names, not fields.
func fieldPaths(prefix string, v any, out map[string]bool) {
	switch x := v.(type) {
	case map[string]any:
		for k, c := range x {
			p := prefix + "." + k
			out[p] = true
			if k != "digest" && !strings.HasSuffix(k, "Digest") {
				fieldPaths(p, c, out)
			}
		}
	case []any:
		for _, c := range x {
			fieldPaths(prefix+"[*]", c, out)
		}
	}
}

func TestExampleIsFullStatementOfOurType(t *testing.T) {
	st := exampleStatement(t, readSpec(t))
	if st["_type"] != "https://in-toto.io/Statement/v1" {
		t.Errorf("_type = %v", st["_type"])
	}
	if st["predicateType"] != typeURI {
		t.Errorf("predicateType = %v, want %s", st["predicateType"], typeURI)
	}
	if subj, _ := st["subject"].([]any); len(subj) == 0 {
		t.Error("example has no subject")
	}
}

// fieldRows returns the field paths of the Fields table rows that have a
// non-empty type, required or optional, and a non-empty example.
func fieldRows(doc string) map[string]bool {
	row := regexp.MustCompile("(?m)^\\| `(predicate[^`]*)` \\| *[^|\\s][^|]* \\| (?:required|optional) \\| *[^|\\s][^|]* \\|")
	rows := map[string]bool{}
	for _, m := range row.FindAllStringSubmatch(doc, -1) {
		rows[m[1]] = true
	}
	return rows
}

func TestEveryExampleFieldIsDocumented(t *testing.T) {
	doc := readSpec(t)
	paths := map[string]bool{}
	fieldPaths("predicate", exampleStatement(t, doc)["predicate"], paths)
	rows := fieldRows(doc)
	var missing, extra []string
	for p := range paths {
		if !rows[p] {
			missing = append(missing, p)
		}
	}
	for r := range rows {
		if !paths[r] {
			extra = append(extra, r)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 {
		t.Errorf("fields in the example without a complete table row (type, required/optional, example): %v", missing)
	}
	if len(extra) > 0 {
		t.Errorf("table rows for fields not in the example: %v", extra)
	}
}

func TestSpecUsesNormativeLanguage(t *testing.T) {
	if n := strings.Count(readSpec(t), "MUST"); n < 5 {
		t.Errorf("spec has %d MUST, want at least 5", n)
	}
}

func TestRelativeLinksResolve(t *testing.T) {
	files := []string{specFile, "README.md", "../docs/decisions/0001-own-predicate.md"}
	link := regexp.MustCompile(`\]\(([^)#\s]+)(#[^)]*)?\)`)
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Errorf("read %s: %v", f, err)
			continue
		}
		for _, m := range link.FindAllStringSubmatch(string(b), -1) {
			target := m[1]
			if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
				continue
			}
			if _, err := os.Stat(filepath.Join(filepath.Dir(f), target)); err != nil {
				t.Errorf("%s: broken link %s", f, target)
			}
		}
	}
}

// The path rule alone passes a clean-named symlink, so the spec must also
// pin how a consumer reads a path: from the git tree, regular files only.
func TestPathRuleCoversSymlinks(t *testing.T) {
	doc := readSpec(t)
	for _, want := range []string{"git tree", "`git ls-tree <commit> -- <path>`", "`100644`", "`100755`", "symlink (`120000`)", "submodule (`160000`)", "raw git blob bytes", "GIT_NO_REPLACE_OBJECTS=1"} {
		if !strings.Contains(doc, want) {
			t.Errorf("parsing rules do not mention %s", want)
		}
	}
}
