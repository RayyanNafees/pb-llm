package formatter

import (
	"fmt"
	"strings"
	"time"

	"pb-llm/internal/types"
)

type Formatter interface {
	FormatCompact(docs []types.DocSection) ([]byte, error)
	FormatText(docs []types.DocSection) ([]byte, error)
}

type CompactFormatter struct{}

func NewCompactFormatter() *CompactFormatter {
	return &CompactFormatter{}
}

func (f *CompactFormatter) FormatCompact(docs []types.DocSection) ([]byte, error) {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# POCKETBASE DOCS|%s|%d sections\n", time.Now().Format("2006-01-02"), len(docs)))

	for i, doc := range docs {
		if !doc.Success {
			continue
		}

		content.WriteString(fmt.Sprintf("\n## %d.%s\n", i+1, doc.Title))

		if doc.APIRoute != "" && doc.Method != "" {
			content.WriteString(fmt.Sprintf("%s %s\n", doc.Method, doc.APIRoute))
		}

		if len(doc.Parameters) > 0 {
			for _, param := range doc.Parameters {
				req := ""
				if param.Required {
					req = "*"
				}
				content.WriteString(fmt.Sprintf("%s%s(%s):%s\n", param.Name, req, param.Type, param.Description))
			}
		}

		if len(doc.Examples) > 0 {
			for key, example := range doc.Examples {
				lang := strings.Split(key, "_")[0]
				cleaned := f.cleanCodeExample(example)
				content.WriteString(fmt.Sprintf("```%s\n%s\n```\n", lang, cleaned))
			}
		}

		if len(doc.ResponseExamples) > 0 {
			for _, resp := range doc.ResponseExamples {
				if resp.Body != "" {
					compacted := f.compactJSON(resp.Body)
					content.WriteString(fmt.Sprintf("Response %d:\n%s\n", resp.StatusCode, compacted))
				}
			}
		}

		cleanContent := f.cleanMainContent(doc.CleanContent)
		if cleanContent != "" {
			content.WriteString(fmt.Sprintf("%s\n", cleanContent))
		}
	}

	return []byte(content.String()), nil
}

func (f *CompactFormatter) FormatText(docs []types.DocSection) ([]byte, error) {
	var content strings.Builder

	content.WriteString("POCKETBASE DOCUMENTATION\n")
	content.WriteString(strings.Repeat("=", 50) + "\n\n")
	content.WriteString(fmt.Sprintf("Generated: %s | Sections: %d\n\n", time.Now().Format("2006-01-02 15:04:05"), len(docs)))

	for i, doc := range docs {
		if !doc.Success {
			continue
		}

		content.WriteString(fmt.Sprintf("[%d] %s\n", i+1, doc.Title))
		content.WriteString(fmt.Sprintf("URL: %s\n", doc.URL))
		content.WriteString(fmt.Sprintf("Category: %s\n", doc.Category))

		if doc.APIRoute != "" && doc.Method != "" {
			content.WriteString(fmt.Sprintf("API: %s %s\n", doc.Method, doc.APIRoute))
		}

		if len(doc.Parameters) > 0 {
			content.WriteString("Parameters:\n")
			for _, param := range doc.Parameters {
				req := ""
				if param.Required {
					req = " (required)"
				}
				content.WriteString(fmt.Sprintf("- %s (%s)%s: %s\n", param.Name, param.Type, req, param.Description))
			}
		}

		cleanContent := f.cleanMainContent(doc.CleanContent)
		if cleanContent != "" {
			content.WriteString("\nContent:\n")
			content.WriteString(cleanContent)
			content.WriteString("\n")
		}

		content.WriteString(strings.Repeat("-", 80) + "\n\n")
	}

	return []byte(content.String()), nil
}

func (f *CompactFormatter) cleanMainContent(content string) string {
	if content == "" {
		return ""
	}

	content = f.removeRepetitiveDescriptions(content)
	content = strings.ReplaceAll(content, "\n\n\n", "\n")
	content = strings.ReplaceAll(content, "\n\n", "\n")
	content = f.removeUINoisePatterns(content)

	lines := strings.Split(content, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if len(line) < 3 || f.isNoiseLine(line) {
			continue
		}

		cleanLines = append(cleanLines, line)
	}

	return strings.Join(cleanLines, "\n")
}

// removeRepetitiveDescriptions removes the repetitive PocketBase description
func (f *CompactFormatter) removeRepetitiveDescriptions(content string) string {
	repetitivePatterns := []string{
		"Open Source backend in 1 file with realtime database, authentication, file storage and admin dashboard",
		"Open source backend in 1 file with realtime database, authentication, file storage and admin dashboard",
		"# Introduction",
		"# How to use PocketBase",
		"# Collections",
		"# Authentication",
		"# Files upload and handling",
	}

	for _, pattern := range repetitivePatterns {
		content = strings.ReplaceAll(content, pattern+"\n", "")
		content = strings.ReplaceAll(content, pattern+"  ", "")

		if strings.HasPrefix(content, pattern) {
			content = strings.TrimPrefix(content, pattern)
			content = strings.TrimLeft(content, "\n ")
		}
	}

	return content
}

func (f *CompactFormatter) removeUINoisePatterns(content string) string {
	noisePatterns := []string{
		"## Content",
		"## Parameters",
		"## Examples",
		"**API Endpoint:**",
		"**METADATA:**",
		"**PARAMETERS:**",
		"**CODE EXAMPLES:**",
		"**API RESPONSES:**",
		"**DOCUMENTATION CONTENT:**",
		"Click here",
		"Read more",
		"Learn more",
		"Edit this page",
		"Improve this page",
		"GitHub releases",
		"Download v",
		"for Linux",
		"for Windows",
		"for macOS",
		"Table of contents",
		"On this page",
		"Jump to",
	}

	for _, pattern := range noisePatterns {
		content = strings.ReplaceAll(content, pattern, "")
	}

	return content
}

func (f *CompactFormatter) isNoiseLine(line string) bool {
	lower := strings.ToLower(line)

	noiseChecks := []string{
		"click here", "read more", "learn more", "see more",
		"edit this page", "improve this page", "feedback",
		"github releases", "changelog", "previous", "next",
		"table of contents", "on this page", "jump to",
		"download", ".zip", ".tar.gz", "localhost:",
		"example.com", "placeholder", "todo", "fixme",
		"breadcrumb", "navigation", "sidebar", "footer",
	}

	for _, check := range noiseChecks {
		if strings.Contains(lower, check) {
			return true
		}
	}

	if strings.ContainsAny(line, "===---***+++") && len(strings.TrimFunc(line, func(r rune) bool {
		return strings.ContainsRune("=*-+_#", r)
	})) < 3 {
		return true
	}

	return false
}

func (f *CompactFormatter) cleanCodeExample(code string) string {
	code = strings.ReplaceAll(code, "\n\n", "\n")
	code = strings.ReplaceAll(code, "\t", " ")

	lines := strings.Split(code, "\n")
	var cleanLines []string
	for _, line := range lines {
		cleanLines = append(cleanLines, strings.TrimRight(line, " \t"))
	}

	return strings.TrimSpace(strings.Join(cleanLines, "\n"))
}

func (f *CompactFormatter) compactJSON(json string) string {
	json = strings.ReplaceAll(json, "\n  ", "\n")
	json = strings.ReplaceAll(json, "  ", " ")
	json = strings.ReplaceAll(json, "\n\n", "\n")
	return strings.TrimSpace(json)
}

func GetFormatter(format string) Formatter {
	return NewCompactFormatter()
}
