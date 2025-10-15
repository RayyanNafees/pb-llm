package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"pb-llm/internal/scraper"
)

func main() {
	var (
		help = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	runScraper()
}

func runScraper() {
	fmt.Println("🚀 PocketBase Documentation Scraper for LLMs")
	fmt.Println("===========================================")

	s := scraper.New()

	docs, err := s.ScrapeAll()
	if err != nil {
		log.Fatalf("❌ Scraping failed: %v", err)
	}

	// Generate timestamp for session directory
	timestamp := time.Now().Format("2006-01-02_15-04-05.000")
	sessionDir := fmt.Sprintf("session_%s", timestamp)

	summaryFile := "summary.txt"

	// Always generate both LLM and TXT formats
	formats := []string{"llm", "txt"}
	extensions := []string{".llm.md", ".txt"}

	fmt.Printf("💾 Saving documentation (LLM + TXT) to: docs/%s/\n", sessionDir)

	for i, format := range formats {
		outputFile := fmt.Sprintf("pocketbase_docs%s", extensions[i])
		if err := s.SaveToFile(docs, sessionDir, outputFile, format); err != nil {
			log.Printf("⚠️ Failed to save %s format: %v", format, err)
		} else {
			fmt.Printf("   ✅ %s\n", outputFile)
		}
	}

	// Generate and save summary
	fmt.Printf("💾 Saving summary to: docs/%s/%s\n", sessionDir, summaryFile)
	if err := s.SaveSummaryToFile(docs, sessionDir, summaryFile); err != nil {
		log.Printf("⚠️ Failed to save summary: %v", err)
	}

	fmt.Printf("\n🎉 Scraping completed successfully!\n")
	fmt.Printf("📁 Session directory: docs/%s/\n", sessionDir)
	fmt.Printf("📄 Documentation: pocketbase_docs.llm.md (LLM-optimized), pocketbase_docs.txt (plain text)\n")
	fmt.Printf("📊 Summary: %s (with token counting & LLM metrics)\n", summaryFile)
	fmt.Printf("🤖 Ready for LLM usage!\n")
}

func printHelp() {
	const helpText = `PocketBase Documentation Scraper for LLM Usage
=============================================

DESCRIPTION:
  Scrapes PocketBase documentation and automatically generates both:
  • LLM-optimized format with token counting and AI-friendly structure
  • Plain text format for general use

USAGE:
  go run cmd/main.go [OPTIONS]

OPTIONS:
  -help
        Show this help message

FEATURES:
  🤖 LLM-optimized output format
  📊 Token counting and estimation
  📈 Context window usage analysis
  🔧 AI training dataset structure
  📝 Comprehensive LLM usage statistics
  📄 Plain text backup format

OUTPUT:
  • pocketbase_docs.llm.md - LLM-optimized documentation
  • pocketbase_docs.txt - Plain text documentation
  • summary.txt - Comprehensive statistics with token analysis

EXAMPLE:
  go run cmd/main.go

All files saved in timestamped docs/session_YYYY-MM-DD_HH-MM-SS.mmm/ directory`

	fmt.Println(helpText)
}
