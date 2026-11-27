package razorpay

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// staleAuthoringGuidePatterns are templates that no longer match the codebase.
var staleAuthoringGuidePatterns = []*regexp.Regexp{
	regexp.MustCompile(`log \*slog\.Logger`),
	regexp.MustCompile(`\(log, client\)`),
	regexp.MustCompile(`Register the tool in ` + "`server.go`"),
}

func TestAuthoringGuide_NoStaleToolSignatures(t *testing.T) {
	lines, err := readREADMELines()
	require.NoError(t, err)

	for lineNum, line := range lines {
		for _, pattern := range staleAuthoringGuidePatterns {
			if pattern.MatchString(line) {
				t.Errorf(
					"README.md:%d: stale authoring guide pattern %q: %s",
					lineNum+1,
					pattern.String(),
					strings.TrimSpace(line),
				)
			}
		}
	}
}

func TestAuthoringGuide_ValidatorReceiverMatchesVariable(t *testing.T) {
	lines, err := readREADMELines()
	require.NoError(t, err)

	inCodeBlock := false
	usesVValidator := false

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			usesVValidator = false
			continue
		}
		if !inCodeBlock {
			continue
		}

		if strings.Contains(line, "v := NewValidator") ||
			strings.Contains(line, "v := NewValidator(&r)") {
			usesVValidator = true
		}

		if strings.Contains(line, "validator.HandleErrorsIfAny()") &&
			usesVValidator {
			t.Errorf(
				"README.md:%d: use v.HandleErrorsIfAny() when receiver is v",
				lineNum+1,
			)
		}
	}
}

func TestAuthoringGuide_DocumentsCurrentToolSignature(t *testing.T) {
	content, err := os.ReadFile("README.md")
	require.NoError(t, err)

	text := string(content)
	assert.Contains(t, text, "obs *observability.Observability")
	assert.Contains(t, text, "v.HandleErrorsIfAny()")
}

func readREADMELines() ([]string, error) {
	file, err := os.Open("README.md")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}
