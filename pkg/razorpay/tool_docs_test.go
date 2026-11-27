package razorpay

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func TestToolGeneratorDocs_requireArrayItemsRule(t *testing.T) {
	root := repoRoot(t)
	docs := []string{
		"AGENTS.md",
		filepath.Join(".cursor", "skills", "razorpay-mcp-tool-gen", "SKILL.md"),
		filepath.Join(".claude", "skills", "razorpay-mcp-tool-gen", "SKILL.md"),
		filepath.Join(".agents", "skills", "razorpay-mcp-tool-gen", "SKILL.md"),
	}

	for _, doc := range docs {
		t.Run(doc, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(root, doc))
			require.NoError(t, err)
			text := string(content)
			assert.Contains(t, text, "mcpgo.Items")
			assert.Contains(t, text, "MUST include `mcpgo.Items")
		})
	}
}

func TestREADME_toolNamesMatchRegisteredTools(t *testing.T) {
	root := repoRoot(t)
	readmePath := filepath.Join(root, "README.md")
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err)

	readmeTools := parseREADMEToolNames(string(content))
	require.NotEmpty(t, readmeTools)

	registered := collectRegisteredToolNames(t, root)
	for _, name := range readmeTools {
		assert.Contains(t, registered, name,
			"README lists tool %q that is not registered in code", name)
	}
}

func parseREADMEToolNames(readme string) []string {
	const section = "## Available Tools"
	start := strings.Index(readme, section)
	if start == -1 {
		return nil
	}

	sectionText := readme[start:]
	if end := strings.Index(sectionText, "\n## Use Cases"); end != -1 {
		sectionText = sectionText[:end]
	}

	re := regexp.MustCompile(`\|\s*` + "`" + `([a-z_]+)` + "`")
	matches := re.FindAllStringSubmatch(sectionText, -1)
	names := make([]string, 0, len(matches))
	for _, match := range matches {
		names = append(names, match[1])
	}
	return names
}

func collectRegisteredToolNames(t *testing.T, root string) map[string]struct{} {
	t.Helper()

	toolDirs := []string{
		filepath.Join(root, "pkg", "razorpay"),
		filepath.Join(root, "pkg", "razorpay", "integrations"),
	}
	re := regexp.MustCompile(`NewTool\(\s*"([a-z_]+)"`)

	registered := make(map[string]struct{})
	for _, dir := range toolDirs {
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") ||
				strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			fileContent, readErr := os.ReadFile(filepath.Join(dir, entry.Name()))
			require.NoError(t, readErr)
			for _, match := range re.FindAllStringSubmatch(string(fileContent), -1) {
				registered[match[1]] = struct{}{}
			}
		}
	}
	return registered
}
