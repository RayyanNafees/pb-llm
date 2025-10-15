package formatter

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"pb-llm/internal/types"
)

type Formatter interface {
	FormatFull(docs []types.DocSection) ([]byte, error)
	FormatSimplified(docs []types.DocSection) ([]byte, error)
}

// LLMFormatter handles AI/LLM optimized output formatting
type LLMFormatter struct {
	tokenEstimator *types.TokenEstimator
}

func NewLLMFormatter() *LLMFormatter {
	return &LLMFormatter{
		tokenEstimator: types.NewTokenEstimator(),
	}
}

func (f *LLMFormatter) FormatFull(docs []types.DocSection) ([]byte, error) {
	var content strings.Builder

	// Header with metadata for LLMs
	content.WriteString("# POCKETBASE DOCUMENTATION - LLM TRAINING DATASET\n\n")
	content.WriteString("## DATASET METADATA\n")
	content.WriteString(fmt.Sprintf("- Generated: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("- Total Sections: %d\n", len(docs)))
	content.WriteString("- Format: Structured for AI/LLM consumption\n")
	content.WriteString("- Source: PocketBase Official Documentation\n")
	content.WriteString("- Purpose: Backend-as-a-Service knowledge base\n\n")

	// Quick reference index
	content.WriteString("## DOCUMENTATION INDEX\n")
	categoryMap := make(map[string][]string)
	for _, doc := range docs {
		if doc.Success {
			category := strings.Title(doc.Category)
			categoryMap[category] = append(categoryMap[category], doc.Title)
		}
	}

	for category, titles := range categoryMap {
		content.WriteString(fmt.Sprintf("### %s (%d sections)\n", category, len(titles)))
		for _, title := range titles {
			content.WriteString(fmt.Sprintf("- %s\n", title))
		}
		content.WriteString("\n")
	}

	content.WriteString("---\n\n")

	// Main content structured for LLM consumption
	for i, doc := range docs {
		if !doc.Success {
			continue
		}

		tokenStats := f.tokenEstimator.EstimateDocTokens(doc)

		content.WriteString(fmt.Sprintf("## SECTION %d: %s\n\n", i+1, doc.Title))

		// Structured metadata
		content.WriteString("**METADATA:**\n")
		content.WriteString(fmt.Sprintf("- URL: %s\n", doc.URL))
		content.WriteString(fmt.Sprintf("- Category: %s\n", doc.Category))
		content.WriteString(fmt.Sprintf("- Estimated Tokens: %d\n", tokenStats.LLMUsable))

		if doc.APIRoute != "" && doc.Method != "" {
			content.WriteString(fmt.Sprintf("- API Endpoint: `%s %s`\n", doc.Method, doc.APIRoute))
		}

		if doc.Description != "" {
			content.WriteString(fmt.Sprintf("- Description: %s\n", doc.Description))
		}
		content.WriteString("\n")

		// Parameters in structured format
		if len(doc.Parameters) > 0 {
			content.WriteString("**PARAMETERS:**\n")
			for _, param := range doc.Parameters {
				required := ""
				if param.Required {
					required = " [REQUIRED]"
				}
				content.WriteString(fmt.Sprintf("- `%s` (%s)%s: %s\n", param.Name, param.Type, required, param.Description))
			}
			content.WriteString("\n")
		}

		// Code examples optimized for LLMs
		if len(doc.Examples) > 0 {
			content.WriteString("**CODE EXAMPLES:**\n")
			for key, example := range doc.Examples {
				lang, _, _ := strings.Cut(key, "_")
				content.WriteString(fmt.Sprintf("```%s\n// %s Example for %s\n%s\n```\n\n", lang, strings.Title(lang), doc.Title, example))
			}
		}

		// Response examples
		if len(doc.ResponseExamples) > 0 {
			content.WriteString("**API RESPONSES:**\n")
			for _, resp := range doc.ResponseExamples {
				content.WriteString(fmt.Sprintf("HTTP %d - %s:\n```json\n%s\n```\n\n", resp.StatusCode, resp.Description, resp.Body))
			}
		}

		// Main content with clear boundaries
		content.WriteString("**DOCUMENTATION CONTENT:**\n")
		content.WriteString("```markdown\n")
		content.WriteString(doc.CleanContent)
		content.WriteString("\n```\n\n")

		content.WriteString("---\n\n")
	}

	// Footer with usage instructions
	content.WriteString("## LLM USAGE INSTRUCTIONS\n\n")
	content.WriteString("This dataset contains comprehensive PocketBase documentation structured for AI consumption.\n")
	content.WriteString("Each section includes:\n")
	content.WriteString("- Structured metadata for context\n")
	content.WriteString("- API parameters and examples\n")
	content.WriteString("- Clean documentation content\n")
	content.WriteString("- Token estimates for processing\n\n")
	content.WriteString("Use this data to answer questions about PocketBase, generate code examples, ")
	content.WriteString("or provide implementation guidance for backend development.\n")

	return []byte(content.String()), nil
}

func (f *LLMFormatter) FormatSimplified(docs []types.DocSection) ([]byte, error) {
	var content strings.Builder

	content.WriteString("# PocketBase Documentation (LLM Simplified)\n\n")
	content.WriteString(fmt.Sprintf("Generated: %s | Sections: %d | Format: AI-Optimized\n\n",
		time.Now().Format("2006-01-02 15:04:05"), len(docs)))

	for i, doc := range docs {
		if !doc.Success {
			continue
		}

		content.WriteString(fmt.Sprintf("## %d. %s\n", i+1, doc.Title))
		content.WriteString(fmt.Sprintf("**Category:** %s | **URL:** %s\n\n", doc.Category, doc.URL))

		if doc.APIRoute != "" && doc.Method != "" {
			content.WriteString(fmt.Sprintf("**API:** `%s %s`\n\n", doc.Method, doc.APIRoute))
		}

		if len(doc.Parameters) > 0 {
			content.WriteString("**Key Parameters:** ")
			var paramNames []string
			for _, param := range doc.Parameters {
				if param.Required {
					paramNames = append(paramNames, fmt.Sprintf("`%s` (required)", param.Name))
				} else {
					paramNames = append(paramNames, fmt.Sprintf("`%s`", param.Name))
				}
			}
			content.WriteString(strings.Join(paramNames, ", ") + "\n\n")
		}

		// Condensed content
		lines := strings.Split(doc.CleanContent, "\n")
		if len(lines) > 10 {
			content.WriteString(strings.Join(lines[:5], "\n") + "\n[...content truncated for brevity...]\n" + strings.Join(lines[len(lines)-3:], "\n") + "\n\n")
		} else {
			content.WriteString(doc.CleanContent + "\n\n")
		}

		content.WriteString("---\n\n")
	}

	return []byte(content.String()), nil
}

// JSONFormatter handles JSON output formatting
type JSONFormatter struct{}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) FormatFull(docs []types.DocSection) ([]byte, error) {
	return json.MarshalIndent(docs, "", "  ")
}

func (f *JSONFormatter) FormatSimplified(docs []types.DocSection) ([]byte, error) {
	simplified := make([]types.SimplifiedDoc, len(docs))
	for i, doc := range docs {
		simplified[i] = types.SimplifiedDoc{
			Title:        doc.Title,
			URL:          doc.URL,
			Category:     doc.Category,
			Description:  doc.Description,
			APIRoute:     doc.APIRoute,
			Method:       doc.Method,
			Parameters:   doc.Parameters,
			Examples:     doc.Examples,
			CleanContent: doc.CleanContent,
		}
	}
	return json.MarshalIndent(simplified, "", "  ")
}

// MarkdownFormatter handles Markdown output formatting
type MarkdownFormatter struct{}

func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

func (f *MarkdownFormatter) FormatFull(docs []types.DocSection) ([]byte, error) {
	var content strings.Builder

	content.WriteString("# PocketBase Documentation\n\n")
	content.WriteString(fmt.Sprintf("Generated on: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString("---\n\n")

	for _, doc := range docs {
		content.WriteString(fmt.Sprintf("## %s\n\n", doc.Title))
		content.WriteString(fmt.Sprintf("**URL:** %s  \n", doc.URL))
		content.WriteString(fmt.Sprintf("**Category:** %s  \n", doc.Category))

		if doc.APIRoute != "" && doc.Method != "" {
			content.WriteString(fmt.Sprintf("**API Endpoint:** `%s %s`  \n", doc.Method, doc.APIRoute))
		}

		if doc.Description != "" {
			content.WriteString(fmt.Sprintf("**Description:** %s  \n", doc.Description))
		}

		content.WriteString("\n")

		if len(doc.Parameters) > 0 {
			content.WriteString("### Parameters\n\n")
			for _, param := range doc.Parameters {
				required := ""
				if param.Required {
					required = " *(required)*"
				}
				content.WriteString(fmt.Sprintf("- **%s** (%s)%s: %s\n", param.Name, param.Type, required, param.Description))
			}
			content.WriteString("\n")
		}

		if len(doc.Examples) > 0 {
			content.WriteString("### Code Examples\n\n")
			for key, example := range doc.Examples {
				lang, _, _ := strings.Cut(key, "_")
				content.WriteString(fmt.Sprintf("#### %s\n\n", strings.Title(lang)))
				content.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", lang, example))
			}
		}

		if len(doc.ResponseExamples) > 0 {
			content.WriteString("### Response Examples\n\n")
			for _, resp := range doc.ResponseExamples {
				content.WriteString(fmt.Sprintf("**HTTP %d** - %s\n\n", resp.StatusCode, resp.Description))
				content.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", resp.Body))
			}
		}

		content.WriteString("### Content\n\n")
		content.WriteString(doc.CleanContent)
		content.WriteString("\n\n---\n\n")
	}

	return []byte(content.String()), nil
}

func (f *MarkdownFormatter) FormatSimplified(docs []types.DocSection) ([]byte, error) {
	var content strings.Builder

	content.WriteString("# PocketBase Documentation (Simplified)\n\n")
	content.WriteString(fmt.Sprintf("Generated on: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString("---\n\n")

	for _, doc := range docs {
		content.WriteString(fmt.Sprintf("## %s\n\n", doc.Title))
		content.WriteString(fmt.Sprintf("**URL:** %s  \n", doc.URL))
		content.WriteString(fmt.Sprintf("**Category:** %s  \n", doc.Category))

		if doc.APIRoute != "" && doc.Method != "" {
			content.WriteString(fmt.Sprintf("**API Endpoint:** `%s %s`  \n", doc.Method, doc.APIRoute))
		}

		if doc.Description != "" {
			content.WriteString(fmt.Sprintf("**Description:** %s  \n", doc.Description))
		}

		if len(doc.Parameters) > 0 {
			content.WriteString("\n**Parameters:**\n")
			for _, param := range doc.Parameters {
				required := ""
				if param.Required {
					required = " *(required)*"
				}
				content.WriteString(fmt.Sprintf("- `%s` (%s)%s: %s\n", param.Name, param.Type, required, param.Description))
			}
		}

		content.WriteString("\n")
		content.WriteString(doc.CleanContent)
		content.WriteString("\n\n---\n\n")
	}

	return []byte(content.String()), nil
}

// TextFormatter handles plain text output formatting
type TextFormatter struct{}

func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

func (f *TextFormatter) FormatFull(docs []types.DocSection) ([]byte, error) {
	var content strings.Builder

	content.WriteString("POCKETBASE DOCUMENTATION\n")
	content.WriteString(strings.Repeat("=", 50) + "\n\n")
	content.WriteString(fmt.Sprintf("Generated on: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	for i, doc := range docs {
		content.WriteString(fmt.Sprintf("[%d] %s\n", i+1, doc.Title))
		content.WriteString(fmt.Sprintf("URL: %s\n", doc.URL))
		content.WriteString(fmt.Sprintf("Category: %s\n", doc.Category))

		if doc.APIRoute != "" && doc.Method != "" {
			content.WriteString(fmt.Sprintf("API Endpoint: %s %s\n", doc.Method, doc.APIRoute))
		}

		if doc.Description != "" {
			content.WriteString(fmt.Sprintf("Description: %s\n", doc.Description))
		}

		if len(doc.Parameters) > 0 {
			content.WriteString("\nParameters:\n")
			for _, param := range doc.Parameters {
				required := ""
				if param.Required {
					required = " (required)"
				}
				content.WriteString(fmt.Sprintf("- %s (%s)%s: %s\n", param.Name, param.Type, required, param.Description))
			}
		}

		if len(doc.Examples) > 0 {
			content.WriteString("\nCode Examples:\n")
			for key, example := range doc.Examples {
				lang, _, _ := strings.Cut(key, "_")
				content.WriteString(fmt.Sprintf("\n%s:\n%s\n", strings.ToUpper(lang), example))
			}
		}

		content.WriteString("\nContent:\n")
		content.WriteString(doc.CleanContent)
		content.WriteString("\n" + strings.Repeat("-", 80) + "\n\n")
	}

	return []byte(content.String()), nil
}

func (f *TextFormatter) FormatSimplified(docs []types.DocSection) ([]byte, error) {
	var content strings.Builder

	content.WriteString("POCKETBASE DOCUMENTATION (SIMPLIFIED)\n")
	content.WriteString(strings.Repeat("=", 50) + "\n\n")
	content.WriteString(fmt.Sprintf("Generated on: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	for i, doc := range docs {
		content.WriteString(fmt.Sprintf("[%d] %s\n", i+1, doc.Title))
		content.WriteString(fmt.Sprintf("URL: %s\n", doc.URL))
		content.WriteString(fmt.Sprintf("Category: %s\n", doc.Category))

		if doc.APIRoute != "" && doc.Method != "" {
			content.WriteString(fmt.Sprintf("API Endpoint: %s %s\n", doc.Method, doc.APIRoute))
		}

		if doc.Description != "" {
			content.WriteString(fmt.Sprintf("Description: %s\n", doc.Description))
		}

		if len(doc.Parameters) > 0 {
			content.WriteString("\nParameters:\n")
			for _, param := range doc.Parameters {
				required := ""
				if param.Required {
					required = " (required)"
				}
				content.WriteString(fmt.Sprintf("- %s (%s)%s: %s\n", param.Name, param.Type, required, param.Description))
			}
		}

		content.WriteString("\nContent:\n")
		content.WriteString(doc.CleanContent)
		content.WriteString("\n" + strings.Repeat("-", 80) + "\n\n")
	}

	return []byte(content.String()), nil
}

// GetFormatter returns appropriate formatter based on format string
func GetFormatter(format string) Formatter {
	switch strings.ToLower(format) {
	case "md", "markdown":
		return NewMarkdownFormatter()
	case "txt", "text":
		return NewTextFormatter()
	case "llm", "ai":
		return NewLLMFormatter()
	default:
		return NewJSONFormatter()
	}
}
