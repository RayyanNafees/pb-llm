package types

import (
	"regexp"
	"strings"
	"unicode"
)

type DocSection struct {
	Title            string            `json:"title"`
	URL              string            `json:"url"`
	Content          string            `json:"content"`
	CleanContent     string            `json:"clean_content"`
	APIRoute         string            `json:"api_route,omitempty"`
	Method           string            `json:"method,omitempty"`
	Parameters       []Parameter       `json:"parameters,omitempty"`
	Examples         map[string]string `json:"examples,omitempty"`
	Description      string            `json:"description"`
	Category         string            `json:"category"`
	Headers          []string          `json:"headers,omitempty"`
	ResponseExamples []ResponseExample `json:"response_examples,omitempty"`
	Success          bool              `json:"success"`
	Error            string            `json:"error,omitempty"`
}

type Parameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}

type ResponseExample struct {
	StatusCode  int    `json:"status_code"`
	Description string `json:"description"`
	Body        string `json:"body"`
}

type SimplifiedDoc struct {
	Title        string            `json:"title"`
	URL          string            `json:"url"`
	Category     string            `json:"category"`
	Description  string            `json:"description"`
	APIRoute     string            `json:"api_route,omitempty"`
	Method       string            `json:"method,omitempty"`
	Parameters   []Parameter       `json:"parameters,omitempty"`
	Examples     map[string]string `json:"examples,omitempty"`
	CleanContent string            `json:"clean_content"`
}

// TokenEstimator provides rough token counting for LLM usage
type TokenEstimator struct {
	wordRegex *regexp.Regexp
}

// DocTokenStats holds token statistics for a document section
type DocTokenStats struct {
	Title        int `json:"title_tokens"`
	Description  int `json:"description_tokens"`
	Content      int `json:"content_tokens"`
	CleanContent int `json:"clean_content_tokens"`
	Parameters   int `json:"parameters_tokens"`
	Examples     int `json:"examples_tokens"`
	Total        int `json:"total_tokens"`
	LLMUsable    int `json:"llm_usable_tokens"`
}

// SummaryStats holds comprehensive statistics for LLM usage
type SummaryStats struct {
	TotalSections       int                      `json:"total_sections"`
	SuccessfulSections  int                      `json:"successful_sections"`
	FailedSections      int                      `json:"failed_sections"`
	TotalTokens         int                      `json:"total_tokens"`
	LLMUsableTokens     int                      `json:"llm_usable_tokens"`
	AvgTokensPerSection int                      `json:"avg_tokens_per_section"`
	CategoryStats       map[string]CategoryStats `json:"category_stats"`
	LargestSection      SectionStat              `json:"largest_section"`
	SmallestSection     SectionStat              `json:"smallest_section"`
}

// CategoryStats holds statistics per category
type CategoryStats struct {
	Count       int `json:"count"`
	TotalTokens int `json:"total_tokens"`
	AvgTokens   int `json:"avg_tokens"`
}

// SectionStat holds basic section statistics
type SectionStat struct {
	Title  string `json:"title"`
	Tokens int    `json:"tokens"`
	Size   int    `json:"size_chars"`
}

func NewTokenEstimator() *TokenEstimator {
	return &TokenEstimator{
		wordRegex: regexp.MustCompile(`\w+`),
	}
}

func (te *TokenEstimator) EstimateTokens(text string) int {
	if text == "" {
		return 0
	}

	words := te.wordRegex.FindAllString(text, -1)
	wordCount := len(words)

	punctCount := 0
	for _, r := range text {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			punctCount++
		}
	}

	tokenEstimate := int(float64(wordCount)*0.75 + float64(punctCount)*0.5)

	if tokenEstimate == 0 && len(strings.TrimSpace(text)) > 0 {
		tokenEstimate = 1
	}

	return tokenEstimate
}

func (te *TokenEstimator) EstimateDocTokens(doc DocSection) DocTokenStats {
	stats := DocTokenStats{
		Title:        te.EstimateTokens(doc.Title),
		Description:  te.EstimateTokens(doc.Description),
		Content:      te.EstimateTokens(doc.Content),
		CleanContent: te.EstimateTokens(doc.CleanContent),
	}

	for _, param := range doc.Parameters {
		stats.Parameters += te.EstimateTokens(param.Name + " " + param.Type + " " + param.Description)
	}

	for _, example := range doc.Examples {
		stats.Examples += te.EstimateTokens(example)
	}

	stats.Total = stats.Title + stats.Description + stats.Content + stats.Parameters + stats.Examples
	stats.LLMUsable = stats.Title + stats.Description + stats.CleanContent + stats.Parameters + stats.Examples

	return stats
}
