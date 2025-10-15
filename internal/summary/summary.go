package summary

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"pb-llm/internal/types"
)

type Generator struct {
	tokenEstimator *types.TokenEstimator
}

func New() *Generator {
	return &Generator{
		tokenEstimator: types.NewTokenEstimator(),
	}
}

func (g *Generator) GenerateReport(docs []types.DocSection) string {
	var summary strings.Builder

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	stats := g.calculateComprehensiveStats(docs)

	summary.WriteString("================================================================================\n")
	summary.WriteString("POCKETBASE DOCUMENTATION SCRAPING SUMMARY FOR LLM USAGE\n")
	summary.WriteString("================================================================================\n")
	summary.WriteString(fmt.Sprintf("Generated: %s\n", timestamp))
	summary.WriteString(fmt.Sprintf("Total Sections: %d\n", len(docs)))
	summary.WriteString("Purpose: LLM Training & Context Enhancement\n")

	summary.WriteString("\n🤖 LLM USAGE STATISTICS\n")
	summary.WriteString("================================================================================\n")
	summary.WriteString(fmt.Sprintf("📊 Total Estimated Tokens: %s\n", g.formatNumber(stats.TotalTokens)))
	summary.WriteString(fmt.Sprintf("🔧 LLM-Ready Tokens: %s\n", g.formatNumber(stats.LLMUsableTokens)))
	summary.WriteString(fmt.Sprintf("📈 Avg Tokens/Section: %s\n", g.formatNumber(stats.AvgTokensPerSection)))
	summary.WriteString(fmt.Sprintf("📖 Context Window Usage (120K): ~%.1f%%\n", float64(stats.LLMUsableTokens)/120000.0*100))
	summary.WriteString(fmt.Sprintf("📖 Context Window Usage (200K): ~%.1f%%\n", float64(stats.LLMUsableTokens)/200000.0*100))
	summary.WriteString(fmt.Sprintf("🎯 Compression Ratio: %.1f%% (clean vs raw)\n", float64(stats.LLMUsableTokens)/float64(stats.TotalTokens)*100))

	summary.WriteString("\n📊 PROCESSING STATISTICS\n")
	summary.WriteString("================================================================================\n")
	summary.WriteString(fmt.Sprintf("✅ Successful: %d/%d sections (%.1f%%)\n", stats.SuccessfulSections, stats.TotalSections, float64(stats.SuccessfulSections)/float64(stats.TotalSections)*100))
	summary.WriteString(fmt.Sprintf("❌ Failed: %d/%d sections (%.1f%%)\n", stats.FailedSections, stats.TotalSections, float64(stats.FailedSections)/float64(stats.TotalSections)*100))

	// Content insights
	summary.WriteString("\n📈 CONTENT INSIGHTS\n")
	summary.WriteString("================================================================================\n")
	summary.WriteString(fmt.Sprintf("📄 Largest Section: %s (%s tokens, %s chars)\n", stats.LargestSection.Title, g.formatNumber(stats.LargestSection.Tokens), g.formatNumber(stats.LargestSection.Size)))
	summary.WriteString(fmt.Sprintf("📄 Smallest Section: %s (%s tokens, %s chars)\n", stats.SmallestSection.Title, g.formatNumber(stats.SmallestSection.Tokens), g.formatNumber(stats.SmallestSection.Size)))

	summary.WriteString("\n📂 CONTENT BY CATEGORY (with Token Analysis)\n")
	summary.WriteString("================================================================================\n")

	type categoryInfo struct {
		name  string
		stats types.CategoryStats
	}
	var categories []categoryInfo
	for name, catStats := range stats.CategoryStats {
		categories = append(categories, categoryInfo{name, catStats})
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].stats.TotalTokens > categories[j].stats.TotalTokens
	})

	for _, cat := range categories {
		tokenPercentage := float64(cat.stats.TotalTokens) / float64(stats.LLMUsableTokens) * 100
		summary.WriteString(fmt.Sprintf("   %s: %d sections, %s tokens (%.1f%%), avg %s tokens/section\n",
			cat.name, cat.stats.Count, g.formatNumber(cat.stats.TotalTokens), tokenPercentage, g.formatNumber(cat.stats.AvgTokens)))
	}

	successful := 0
	failed := 0
	totalContent := 0
	totalParams := 0
	methodCount := make(map[string]int)

	var failedSections []string
	var successfulSections []string

	for i, doc := range docs {
		if doc.Success {
			successful++
			tokenStats := g.tokenEstimator.EstimateDocTokens(doc)
			successfulSections = append(successfulSections, fmt.Sprintf("[%d] %s (%s tokens, %d chars, %d params)",
				i+1, doc.Title, g.formatNumber(tokenStats.LLMUsable), len(doc.Content), len(doc.Parameters)))
			totalContent += len(doc.Content)
			totalParams += len(doc.Parameters)
		} else {
			failed++
			errorMsg := doc.Error
			if errorMsg == "" {
				errorMsg = "Unknown error"
			}
			failedSections = append(failedSections, fmt.Sprintf("[%d] %s - Error: %s", i+1, doc.Title, errorMsg))
		}

		if doc.Method != "" {
			methodCount[doc.Method]++
		}
	}

	summary.WriteString("\n📊 LEGACY STATISTICS\n")
	summary.WriteString("================================================================================\n")
	summary.WriteString(fmt.Sprintf("📝 Total content: %s characters\n", g.formatNumber(totalContent)))
	summary.WriteString(fmt.Sprintf("🔧 Total parameters: %d\n", totalParams))

	if len(methodCount) > 0 {
		summary.WriteString("\n🌐 API METHODS FOUND\n")
		summary.WriteString("================================================================================\n")
		for method, count := range methodCount {
			summary.WriteString(fmt.Sprintf("   %s: %d endpoints\n", method, count))
		}
	}

	summary.WriteString("\n✅ SUCCESSFUL SECTIONS\n")
	summary.WriteString("================================================================================\n")
	for _, section := range successfulSections {
		summary.WriteString(section + "\n")
	}

	if len(failedSections) > 0 {
		summary.WriteString("\n❌ FAILED SECTIONS\n")
		summary.WriteString("================================================================================\n")
		for _, section := range failedSections {
			summary.WriteString(section + "\n")
		}
	}

	customPackages := 0
	summary.WriteString("\n🔧 CUSTOM PACKAGES INTEGRATED\n")
	summary.WriteString("================================================================================\n")
	for _, doc := range docs {
		if strings.Contains(doc.URL, "githubusercontent.com") {
			customPackages++
			summary.WriteString("✅ " + doc.Title + " - " + doc.URL + "\n")
		}
	}
	switch customPackages {
	case 0:
		summary.WriteString("No custom packages found.\n")
	default:
		summary.WriteString(fmt.Sprintf("\nTotal custom packages integrated: %d\n", customPackages))
	}

	summary.WriteString("\n================================================================================\n")
	summary.WriteString("END OF SUMMARY\n")
	summary.WriteString("================================================================================\n")

	return summary.String()
}

func (g *Generator) calculateComprehensiveStats(docs []types.DocSection) types.SummaryStats {
	stats := types.SummaryStats{
		TotalSections:   len(docs),
		CategoryStats:   make(map[string]types.CategoryStats),
		LargestSection:  types.SectionStat{Title: "None", Tokens: 0, Size: 0},
		SmallestSection: types.SectionStat{Title: "None", Tokens: 999999, Size: 999999},
	}

	totalTokens := 0
	llmUsableTokens := 0

	categoryTokens := make(map[string]int)
	categoryCounts := make(map[string]int)

	for _, doc := range docs {
		if doc.Success {
			stats.SuccessfulSections++
			tokenStats := g.tokenEstimator.EstimateDocTokens(doc)
			totalTokens += tokenStats.Total
			llmUsableTokens += tokenStats.LLMUsable

			if tokenStats.LLMUsable > stats.LargestSection.Tokens {
				stats.LargestSection = types.SectionStat{
					Title:  doc.Title,
					Tokens: tokenStats.LLMUsable,
					Size:   len(doc.CleanContent),
				}
			}
			if tokenStats.LLMUsable < stats.SmallestSection.Tokens {
				stats.SmallestSection = types.SectionStat{
					Title:  doc.Title,
					Tokens: tokenStats.LLMUsable,
					Size:   len(doc.CleanContent),
				}
			}

			category := strings.Title(doc.Category)
			categoryTokens[category] += tokenStats.LLMUsable
			categoryCounts[category]++
		} else {
			stats.FailedSections++
		}
	}

	stats.TotalTokens = totalTokens
	stats.LLMUsableTokens = llmUsableTokens
	if stats.SuccessfulSections > 0 {
		stats.AvgTokensPerSection = llmUsableTokens / stats.SuccessfulSections
	}

	for category, count := range categoryCounts {
		tokens := categoryTokens[category]
		avgTokens := 0
		if count > 0 {
			avgTokens = tokens / count
		}
		stats.CategoryStats[category] = types.CategoryStats{
			Count:       count,
			TotalTokens: tokens,
			AvgTokens:   avgTokens,
		}
	}

	return stats
}

func (g *Generator) formatNumber(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(char)
	}
	return result.String()
}

func (g *Generator) GetStats(docs []types.DocSection) (successful, failed, totalContent, totalParams int) {
	for _, doc := range docs {
		if doc.Success {
			successful++
			totalContent += len(doc.Content)
			totalParams += len(doc.Parameters)
		} else {
			failed++
		}
	}
	return
}

// GetTokenStats returns comprehensive token statistics
func (g *Generator) GetTokenStats(docs []types.DocSection) types.SummaryStats {
	return g.calculateComprehensiveStats(docs)
}

// GetCategoryBreakdown returns a breakdown of sections by category
func (g *Generator) GetCategoryBreakdown(docs []types.DocSection) map[string]int {
	categoryCount := make(map[string]int)
	for _, doc := range docs {
		if doc.Success {
			categoryCount[strings.Title(doc.Category)]++
		}
	}
	return categoryCount
}
