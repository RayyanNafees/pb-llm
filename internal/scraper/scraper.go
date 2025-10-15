package scraper

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"pb-llm/internal/formatter"
	"pb-llm/internal/summary"
	"pb-llm/internal/types"
)

const (
	BaseURL        = "https://pocketbase.io/docs/"
	RateLimitDelay = 1 * time.Second
	MaxRetries     = 3
	Timeout        = 30 * time.Second
)

var docSections = []types.DocSection{
	{Title: "Introduction", URL: "https://pocketbase.io/docs/", Category: "general"},
	{Title: "How to use PocketBase", URL: "https://pocketbase.io/docs/how-to-use/", Category: "general"},
	{Title: "pb-ext - Enhanced PocketBase Server", URL: "https://raw.githubusercontent.com/magooney-loon/pb-ext/main/README.md", Category: "general"},
	{Title: "pb-ext - Scripts Documentation", URL: "https://raw.githubusercontent.com/magooney-loon/pb-ext/refs/heads/main/cmd/scripts/README.md", Category: "general"},
	{Title: "pb-ext - Collections Implementation", URL: "https://raw.githubusercontent.com/magooney-loon/pb-ext/refs/heads/main/cmd/server/collections.go", Category: "hooks"},
	{Title: "pb-ext - Handlers Implementation", URL: "https://raw.githubusercontent.com/magooney-loon/pb-ext/refs/heads/main/cmd/server/handlers.go", Category: "hooks"},
	{Title: "pb-ext - Jobs Implementation", URL: "https://raw.githubusercontent.com/magooney-loon/pb-ext/refs/heads/main/cmd/server/jobs.go", Category: "hooks"},
	{Title: "pb-ext - Routes Implementation", URL: "https://raw.githubusercontent.com/magooney-loon/pb-ext/refs/heads/main/cmd/server/routes.go", Category: "hooks"},
	{Title: "Going to production", URL: "https://pocketbase.io/docs/going-to-production/", Category: "setup"},
	{Title: "pb-deployer - PocketBase Production Deployment", URL: "https://raw.githubusercontent.com/magooney-loon/pb-deployer/main/README.md", Category: "setup"},
	{Title: "Collections", URL: "https://pocketbase.io/docs/collections/", Category: "database"},
	{Title: "API rules and filters", URL: "https://pocketbase.io/docs/api-rules-and-filters/", Category: "api"},
	{Title: "Working with relations", URL: "https://pocketbase.io/docs/working-with-relations/", Category: "database"},
	{Title: "Authentication", URL: "https://pocketbase.io/docs/authentication/", Category: "auth"},
	{Title: "Files upload and handling", URL: "https://pocketbase.io/docs/files-handling/", Category: "files"},

	{Title: "Collection API", URL: "https://pocketbase.io/docs/api-collections/", Category: "api"},
	{Title: "Record CRUD API", URL: "https://pocketbase.io/docs/api-records/", Category: "api"},
	{Title: "Realtime API", URL: "https://pocketbase.io/docs/api-realtime/", Category: "api"},

	{Title: "Files API", URL: "https://pocketbase.io/docs/api-files/", Category: "api"},
	{Title: "API Crons", URL: "https://pocketbase.io/docs/api-crons/", Category: "api"},
	{Title: "Settings API", URL: "https://pocketbase.io/docs/api-settings/", Category: "api"},
	{Title: "Logs API", URL: "https://pocketbase.io/docs/api-logs/", Category: "api"},
	{Title: "Health API", URL: "https://pocketbase.io/docs/api-health/", Category: "api"},
	{Title: "Backups API", URL: "https://pocketbase.io/docs/api-backups/", Category: "api"},
	{Title: "JavaScript SDK", URL: "https://pocketbase.io/docs/js-overview/", Category: "sdk"},

	// Go Extensions
	{Title: "Go Overview", URL: "https://pocketbase.io/docs/go-overview/", Category: "hooks"},
	{Title: "Go Event hooks", URL: "https://pocketbase.io/docs/go-event-hooks/", Category: "hooks"},
	{Title: "Go Routing", URL: "https://pocketbase.io/docs/go-routing/", Category: "hooks"},
	{Title: "Go Database", URL: "https://pocketbase.io/docs/go-database/", Category: "hooks"},
	{Title: "Go Record operations", URL: "https://pocketbase.io/docs/go-records/", Category: "hooks"},
	{Title: "Go Collection operations", URL: "https://pocketbase.io/docs/go-collections/", Category: "hooks"},
	{Title: "Go migrations", URL: "https://pocketbase.io/docs/go-migrations/", Category: "hooks"},
	{Title: "Go Jobs scheduling", URL: "https://pocketbase.io/docs/go-jobs-scheduling/", Category: "hooks"},
	{Title: "Go Sending emails", URL: "https://pocketbase.io/docs/go-sending-emails/", Category: "hooks"},
	{Title: "Go Rendering templates", URL: "https://pocketbase.io/docs/go-rendering-templates/", Category: "hooks"},
	{Title: "Go Console commands", URL: "https://pocketbase.io/docs/go-console-commands/", Category: "hooks"},
	{Title: "Go Realtime messaging", URL: "https://pocketbase.io/docs/go-realtime/", Category: "hooks"},
	{Title: "Go Filesystem", URL: "https://pocketbase.io/docs/go-filesystem/", Category: "hooks"},
	{Title: "Go Logging", URL: "https://pocketbase.io/docs/go-logging/", Category: "hooks"},
	{Title: "Go Testing", URL: "https://pocketbase.io/docs/go-testing/", Category: "hooks"},
	{Title: "Go Miscellaneous", URL: "https://pocketbase.io/docs/go-miscellaneous/", Category: "hooks"},
	{Title: "Go Record proxy", URL: "https://pocketbase.io/docs/go-record-proxy/", Category: "hooks"},

	// JavaScript Extensions
	{Title: "JavaScript Overview", URL: "https://pocketbase.io/docs/js-overview/", Category: "hooks"},
	{Title: "JavaScript Event hooks", URL: "https://pocketbase.io/docs/js-event-hooks/", Category: "hooks"},
	{Title: "JavaScript Routing", URL: "https://pocketbase.io/docs/js-routing/", Category: "hooks"},
	{Title: "JavaScript Database", URL: "https://pocketbase.io/docs/js-database/", Category: "hooks"},
	{Title: "JavaScript Record operations", URL: "https://pocketbase.io/docs/js-records/", Category: "hooks"},
	{Title: "JavaScript Collection operations", URL: "https://pocketbase.io/docs/js-collections/", Category: "hooks"},
	{Title: "JavaScript migrations", URL: "https://pocketbase.io/docs/js-migrations/", Category: "hooks"},
	{Title: "JavaScript Jobs scheduling", URL: "https://pocketbase.io/docs/js-jobs-scheduling/", Category: "hooks"},
	{Title: "JavaScript Sending emails", URL: "https://pocketbase.io/docs/js-sending-emails/", Category: "hooks"},
	{Title: "JavaScript Rendering templates", URL: "https://pocketbase.io/docs/js-rendering-templates/", Category: "hooks"},
	{Title: "JavaScript Console commands", URL: "https://pocketbase.io/docs/js-console-commands/", Category: "hooks"},
	{Title: "JavaScript Realtime messaging", URL: "https://pocketbase.io/docs/js-realtime/", Category: "hooks"},
	{Title: "JavaScript Filesystem", URL: "https://pocketbase.io/docs/js-filesystem/", Category: "hooks"},
	{Title: "JavaScript Logging", URL: "https://pocketbase.io/docs/js-logging/", Category: "hooks"},
}

type Scraper struct {
	client *http.Client
}

func New() *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: Timeout,
		},
	}
}

func (s *Scraper) ScrapeAll() ([]types.DocSection, error) {
	fmt.Printf("🚀 Starting PocketBase documentation scraping...\n")
	fmt.Printf("📝 Processing %d sections\n\n", len(docSections))

	var results []types.DocSection

	for i, section := range docSections {
		fmt.Printf("⏳ [%d/%d] Processing: %s\n", i+1, len(docSections), section.Title)
		fmt.Printf("🔗 URL: %s\n", section.URL)

		processedSection, err := s.processSection(section)
		if err != nil {
			log.Printf("❌ Error processing %s: %v", section.Title, err)
			processedSection.Success = false
			processedSection.Error = err.Error()
		} else {
			processedSection.Success = true
			fmt.Printf("✅ Successfully processed %s (%d chars)\n", section.Title, len(processedSection.CleanContent))
		}

		results = append(results, processedSection)
		fmt.Printf("💤 Waiting %v before next request...\n\n", RateLimitDelay)
		time.Sleep(RateLimitDelay)
	}

	return results, nil
}

func (s *Scraper) processSection(section types.DocSection) (types.DocSection, error) {
	content, err := s.fetchPageContentWithRetry(section.URL)
	if err != nil {
		return section, fmt.Errorf("failed to fetch page: %w", err)
	}

	// Check content type based on URL
	if strings.Contains(section.URL, "raw.githubusercontent.com") {
		if strings.HasSuffix(section.URL, "README.md") || strings.HasSuffix(section.URL, ".md") {
			s.extractMarkdownContent(&section, content)
		} else if strings.HasSuffix(section.URL, ".go") || strings.HasSuffix(section.URL, ".js") ||
			strings.HasSuffix(section.URL, ".ts") || strings.HasSuffix(section.URL, ".py") {
			s.extractSourceCodeContent(&section, content)
		} else {
			s.extractAllContent(&section, content)
		}
	} else {
		s.extractAllContent(&section, content)
	}
	return section, nil
}

func (s *Scraper) fetchPageContentWithRetry(url string) (string, error) {
	var lastErr error

	for attempt := 1; attempt <= MaxRetries; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PocketBase-Docs-Parser/3.0)")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("Cache-Control", "no-cache")

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d failed: %w", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				lastErr = fmt.Errorf("failed to read body on attempt %d: %w", attempt, err)
				continue
			}
			return string(body), nil
		}

		lastErr = fmt.Errorf("HTTP %d: %s (attempt %d)", resp.StatusCode, resp.Status, attempt)
		if resp.StatusCode == 404 {
			break // Don't retry 404s
		}
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	return "", lastErr
}

func (s *Scraper) extractSourceCodeContent(doc *types.DocSection, content string) {
	// For source code files, wrap content in code blocks
	doc.Content = content

	// Determine language from file extension
	language := "text"
	if strings.HasSuffix(doc.URL, ".go") {
		language = "go"
	} else if strings.HasSuffix(doc.URL, ".js") {
		language = "javascript"
	} else if strings.HasSuffix(doc.URL, ".ts") {
		language = "typescript"
	} else if strings.HasSuffix(doc.URL, ".py") {
		language = "python"
	}

	// Create clean content with proper code formatting
	doc.CleanContent = fmt.Sprintf("```%s\n%s\n```", language, content)

	// Extract description from comments at the top
	lines := strings.Split(content, "\n")
	var description strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "//") {
			// Go-style comments
			desc := strings.TrimSpace(strings.TrimPrefix(line, "//"))
			if desc != "" {
				description.WriteString(desc)
				description.WriteString(" ")
			}
		} else if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "#!") {
			// Python/shell-style comments (but not shebang)
			desc := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if desc != "" {
				description.WriteString(desc)
				description.WriteString(" ")
			}
		} else if strings.HasPrefix(line, "/*") {
			// Multi-line comment start
			desc := strings.TrimSpace(strings.TrimPrefix(line, "/*"))
			desc = strings.TrimSuffix(desc, "*/")
			if desc != "" {
				description.WriteString(desc)
				description.WriteString(" ")
			}
		} else if line != "" && !strings.HasPrefix(line, "package") && !strings.HasPrefix(line, "import") {
			// Stop at first non-comment, non-package, non-import line
			break
		}

		if len(description.String()) > 200 {
			break
		}
	}

	doc.Description = strings.TrimSpace(description.String())
	if len(doc.Description) > 200 {
		doc.Description = doc.Description[:200] + "..."
	}

	// Extract function/type definitions as headers
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "func ") {
			// Go functions
			if idx := strings.Index(line, "("); idx > 5 {
				funcName := strings.TrimSpace(line[5:idx])
				doc.Headers = append(doc.Headers, fmt.Sprintf("Line %d: func %s", i+1, funcName))
			}
		} else if strings.HasPrefix(line, "type ") {
			// Go types
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				doc.Headers = append(doc.Headers, fmt.Sprintf("Line %d: type %s", i+1, parts[1]))
			}
		}
	}

	// Mark as successful
	doc.Success = true
}

func (s *Scraper) extractMarkdownContent(doc *types.DocSection, content string) {
	// For Markdown files, the content is already clean
	doc.Content = content
	doc.CleanContent = content

	// Extract description from first paragraph after title
	lines := strings.Split(content, "\n")
	var description strings.Builder
	titleFound := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			titleFound = true
			continue
		}
		if titleFound && line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "!") && !strings.HasPrefix(line, "[") {
			description.WriteString(line)
			if len(description.String()) > 200 {
				break
			}
			description.WriteString(" ")
		}
	}

	doc.Description = strings.TrimSpace(description.String())
	if len(doc.Description) > 200 {
		doc.Description = doc.Description[:200] + "..."
	}

	// Extract headers for navigation
	doc.Headers = s.extractMarkdownHeaders(content)

	// Mark as successful
	doc.Success = true
}

func (s *Scraper) extractAllContent(doc *types.DocSection, content string) {
	// Extract title - try multiple patterns
	doc.Title = s.extractTitle(content, doc.Title)

	// Extract main content
	doc.Content = s.extractMainContent(content)

	// Extract description
	doc.Description = s.extractDescription(content)

	// Extract API information
	doc.APIRoute, doc.Method = s.extractAPIInfo(content)

	// Extract parameters from tables
	doc.Parameters = s.extractParameters(content)

	// Extract code examples
	doc.Examples = s.extractCodeExamples(content)

	// Extract headers for navigation
	doc.Headers = s.extractHeaders(content)

	// Extract response examples
	doc.ResponseExamples = s.extractResponseExamples(content)

	// Create clean, formatted content
	doc.CleanContent = s.createCleanContent(*doc)
}

func (s *Scraper) extractTitle(content, fallbackTitle string) string {
	// Try to extract h1 title from the main content area
	titlePatterns := []string{
		`<h1[^>]*class="[^"]*title[^"]*"[^>]*>([^<]+)</h1>`,
		`<h1[^>]*>([^<]+)</h1>`,
		`<title>([^<]*?)\s*-\s*[Dd]ocs`,
		`<title>([^<]*)</title>`,
	}

	for _, pattern := range titlePatterns {
		regex := regexp.MustCompile(pattern)
		if matches := regex.FindStringSubmatch(content); len(matches) > 1 {
			title := s.cleanText(matches[1])
			if len(title) > 3 && !strings.Contains(strings.ToLower(title), "pocketbase") {
				return title
			}
		}
	}

	return fallbackTitle
}

func (s *Scraper) extractMainContent(content string) string {
	var extractedContent strings.Builder

	// Extract all text content from the main documentation area
	// PocketBase docs are typically in a main content area or article

	// Remove script and style tags first
	content = s.removeScriptsAndStyles(content)

	// Look for main content containers
	contentPatterns := []string{
		`<main[^>]*>(.*?)</main>`,
		`<article[^>]*>(.*?)</article>`,
		`<div[^>]*class="[^"]*(?:content|main|docs|prose)[^"]*"[^>]*>(.*?)</div>`,
		`<div[^>]*id="[^"]*(?:content|main|docs)[^"]*"[^>]*>(.*?)</div>`,
	}

	var mainContent string
	for _, pattern := range contentPatterns {
		regex := regexp.MustCompile(`(?s)` + pattern)
		if matches := regex.FindStringSubmatch(content); len(matches) > 1 {
			mainContent = matches[1]
			break
		}
	}

	// If no main container found, extract from body
	if mainContent == "" {
		bodyRegex := regexp.MustCompile(`(?s)<body[^>]*>(.*?)</body>`)
		if matches := bodyRegex.FindStringSubmatch(content); len(matches) > 1 {
			mainContent = matches[1]
		} else {
			mainContent = content
		}
	}

	// Extract text from common HTML elements
	elements := []string{"h1", "h2", "h3", "h4", "h5", "h6", "p", "li", "blockquote", "td", "th"}

	for _, element := range elements {
		pattern := fmt.Sprintf(`<%s[^>]*>(.*?)</%s>`, element, element)
		regex := regexp.MustCompile(`(?s)` + pattern)
		matches := regex.FindAllStringSubmatch(mainContent, -1)

		for _, match := range matches {
			if len(match) > 1 {
				text := s.cleanText(match[1])
				if len(text) > 10 { // Only include substantial text
					extractedContent.WriteString(text)
					extractedContent.WriteString("\n\n")
				}
			}
		}
	}

	// Also extract plain text that might not be in standard tags
	// Remove all HTML tags and extract remaining text
	plainTextRegex := regexp.MustCompile(`>([^<]{20,})<`)
	plainMatches := plainTextRegex.FindAllStringSubmatch(mainContent, -1)
	for _, match := range plainMatches {
		if len(match) > 1 {
			text := s.cleanText(match[1])
			if len(text) > 20 && !strings.Contains(strings.ToLower(text), "script") {
				extractedContent.WriteString(text)
				extractedContent.WriteString("\n")
			}
		}
	}

	return s.cleanText(extractedContent.String())
}

func (s *Scraper) removeScriptsAndStyles(content string) string {
	// Remove script tags
	scriptRegex := regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
	content = scriptRegex.ReplaceAllString(content, "")

	// Remove style tags
	styleRegex := regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`)
	content = styleRegex.ReplaceAllString(content, "")

	return content
}

func (s *Scraper) extractDescription(content string) string {
	// Try meta description first
	metaPattern := `<meta[^>]*name="description"[^>]*content="([^"]*)"[^>]*>`
	if matches := regexp.MustCompile(metaPattern).FindStringSubmatch(content); len(matches) > 1 {
		desc := s.cleanText(matches[1])
		if len(desc) > 10 {
			return desc
		}
	}

	// Try first substantial paragraph
	pPattern := `<p[^>]*>([^<]{50,}?)</p>`
	if matches := regexp.MustCompile(pPattern).FindStringSubmatch(content); len(matches) > 1 {
		desc := s.cleanText(matches[1])
		if len(desc) > 20 {
			return desc
		}
	}

	return ""
}

func (s *Scraper) extractAPIInfo(content string) (string, string) {
	// Look for API endpoint patterns in code blocks and text
	patterns := []string{
		`<code[^>]*>(GET|POST|PUT|DELETE|PATCH)\s+([^\s<]+)</code>`,
		`(GET|POST|PUT|DELETE|PATCH)\s+([^\s\n<>]+(?:/[^\s\n<>]+)*)`,
		`<strong[^>]*>(GET|POST|PUT|DELETE|PATCH)</strong>[^<]*<code[^>]*>([^<]+)</code>`,
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		if matches := regex.FindStringSubmatch(content); len(matches) >= 3 {
			method := strings.ToUpper(strings.TrimSpace(matches[1]))
			route := strings.TrimSpace(matches[2])

			// Clean up the route
			if strings.HasPrefix(route, "/") && len(route) > 1 {
				return route, method
			}
		}
	}

	return "", ""
}

func (s *Scraper) extractParameters(content string) []types.Parameter {
	var parameters []types.Parameter

	// Look for parameter tables with improved patterns
	tablePattern := `(?s)<table[^>]*>.*?</table>`
	tableRegex := regexp.MustCompile(tablePattern)
	tables := tableRegex.FindAllString(content, -1)

	for _, table := range tables {
		// Check if this table contains parameters
		if strings.Contains(strings.ToLower(table), "param") ||
			strings.Contains(strings.ToLower(table), "field") ||
			strings.Contains(strings.ToLower(table), "property") {

			// Extract table rows
			rowPattern := `(?s)<tr[^>]*>(.*?)</tr>`
			rowRegex := regexp.MustCompile(rowPattern)
			rows := rowRegex.FindAllStringSubmatch(table, -1)

			for i, row := range rows {
				if i == 0 { // Skip header row typically
					continue
				}
				if len(row) > 1 {
					cells := s.extractTableCells(row[1])
					if len(cells) >= 2 {
						param := types.Parameter{
							Name:     s.cleanText(cells[0]),
							Type:     s.cleanText(cells[1]),
							Required: false,
						}
						if len(cells) > 2 {
							param.Description = s.cleanText(cells[2])
							// Check if required is mentioned
							desc := strings.ToLower(param.Description)
							param.Required = strings.Contains(desc, "required") ||
								strings.Contains(strings.ToLower(param.Type), "required")
						}
						if len(cells) > 3 {
							param.Default = s.cleanText(cells[3])
						}

						if param.Name != "" && param.Type != "" {
							parameters = append(parameters, param)
						}
					}
				}
			}
		}
	}

	return parameters
}

func (s *Scraper) extractTableCells(rowContent string) []string {
	cellPattern := `(?s)<t[dh][^>]*>(.*?)</t[dh]>`
	cellRegex := regexp.MustCompile(cellPattern)
	matches := cellRegex.FindAllStringSubmatch(rowContent, -1)

	var cells []string
	for _, match := range matches {
		if len(match) > 1 {
			// Remove inner HTML tags and clean text
			cellText := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(match[1], "")
			cleaned := s.cleanText(cellText)
			cells = append(cells, cleaned)
		}
	}

	return cells
}

func (s *Scraper) extractCodeExamples(content string) map[string]string {
	examples := make(map[string]string)

	// Extract code blocks with language specification
	codePattern := `(?s)<pre[^>]*><code[^>]*(?:class="[^"]*language-([^"]*)"[^>]*)?>(.*?)</code></pre>`
	codeRegex := regexp.MustCompile(codePattern)
	matches := codeRegex.FindAllStringSubmatch(content, -1)

	for i, match := range matches {
		if len(match) >= 3 {
			lang := match[1]
			code := s.cleanCode(match[2])

			// Detect language if not specified
			if lang == "" || lang == "text" {
				lang = s.detectLanguage(code)
			}

			if len(code) > 10 { // Only include substantial code blocks
				key := fmt.Sprintf("%s_example_%d", lang, i)
				examples[key] = code
			}
		}
	}

	// Also look for simple pre blocks without code tags
	simpleCodePattern := `(?s)<pre[^>]*>(.*?)</pre>`
	simpleCodeRegex := regexp.MustCompile(simpleCodePattern)
	simpleMatches := simpleCodeRegex.FindAllStringSubmatch(content, -1)

	for _, match := range simpleMatches {
		if len(match) > 1 && !strings.Contains(match[0], "<code") {
			code := s.cleanCode(match[1])
			if len(code) > 20 { // Only include substantial code blocks
				lang := s.detectLanguage(code)
				key := fmt.Sprintf("%s_simple_%d", lang, len(examples))
				examples[key] = code
			}
		}
	}

	return examples
}

func (s *Scraper) detectLanguage(code string) string {
	code = strings.ToLower(code)

	if strings.Contains(code, "import pocketbase") || strings.Contains(code, "const pb = new pocketbase") {
		return "javascript"
	} else if strings.Contains(code, "curl") || strings.Contains(code, "wget") || strings.Contains(code, "http get") {
		return "bash"
	} else if strings.Contains(code, "package:pocketbase") || strings.Contains(code, "final pb = pocketbase") {
		return "dart"
	} else if strings.Contains(code, "package main") || strings.Contains(code, "func ") || strings.Contains(code, "import (") {
		return "go"
	} else if strings.Contains(code, "def ") || strings.Contains(code, "import ") && strings.Contains(code, "requests") {
		return "python"
	} else if (strings.Contains(code, "{") && strings.Contains(code, "}")) || strings.Contains(code, "\"id\":") {
		return "json"
	} else if strings.Contains(code, "select ") || strings.Contains(code, "insert ") || strings.Contains(code, "update ") {
		return "sql"
	}

	return "text"
}

func (s *Scraper) extractMarkdownHeaders(content string) []string {
	var headers []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			// Remove markdown header syntax and clean up
			header := strings.TrimSpace(strings.TrimLeft(line, "#"))
			if header != "" {
				headers = append(headers, header)
			}
		}
	}

	return headers
}

func (s *Scraper) extractHeaders(content string) []string {
	var headers []string
	headerPattern := `<h([1-6])[^>]*>([^<]+)</h[1-6]>`
	headerRegex := regexp.MustCompile(headerPattern)
	matches := headerRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			header := s.cleanText(match[2])
			if header != "" && len(header) > 2 {
				headers = append(headers, header)
			}
		}
	}

	return headers
}

func (s *Scraper) extractResponseExamples(content string) []types.ResponseExample {
	var examples []types.ResponseExample

	// Look for JSON response examples in pre/code blocks
	jsonPattern := `(?s)<(?:pre|code)[^>]*>(\{.*?\})</(?:pre|code)>`
	jsonRegex := regexp.MustCompile(jsonPattern)
	matches := jsonRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			body := s.cleanCode(match[1])
			if strings.Contains(body, "{") && len(body) > 20 {
				example := types.ResponseExample{
					StatusCode:  200, // Default
					Description: "API Response",
					Body:        body,
				}
				examples = append(examples, example)
			}
		}
	}

	return examples
}

func (s *Scraper) createCleanContent(section types.DocSection) string {
	var buffer bytes.Buffer

	// Start with title
	buffer.WriteString(fmt.Sprintf("# %s\n\n", section.Title))

	// Add description if available
	if section.Description != "" {
		buffer.WriteString(fmt.Sprintf("%s\n\n", section.Description))
	}

	// Add API endpoint info if available
	if section.APIRoute != "" && section.Method != "" {
		buffer.WriteString(fmt.Sprintf("**API Endpoint:** `%s %s`\n\n", section.Method, section.APIRoute))
	}

	// Add parameters section if available
	if len(section.Parameters) > 0 {
		buffer.WriteString("## Parameters\n\n")
		for _, param := range section.Parameters {
			required := ""
			if param.Required {
				required = " (required)"
			}
			buffer.WriteString(fmt.Sprintf("- **%s** (%s)%s: %s\n", param.Name, param.Type, required, param.Description))
		}
		buffer.WriteString("\n")
	}

	// Add main content
	if section.Content != "" {
		buffer.WriteString("## Content\n\n")
		buffer.WriteString(section.Content)
		buffer.WriteString("\n\n")
	}

	// Add code examples if available
	if len(section.Examples) > 0 {
		buffer.WriteString("## Examples\n\n")
		for key, example := range section.Examples {
			lang := strings.Split(key, "_")[0]
			buffer.WriteString(fmt.Sprintf("### %s\n\n```%s\n%s\n```\n\n", strings.Title(lang), lang, example))
		}
	}

	return buffer.String()
}

func (s *Scraper) cleanText(text string) string {
	// Remove HTML entities
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&#x27;", "'")
	text = strings.ReplaceAll(text, "&#x2F;", "/")

	// Clean up whitespace
	text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// Split into lines and clean each line
	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && len(line) > 1 {
			cleaned = append(cleaned, line)
		}
	}

	result := strings.Join(cleaned, "\n")

	// Remove excessive newlines
	re := regexp.MustCompile(`\n{3,}`)
	result = re.ReplaceAllString(result, "\n\n")

	return strings.TrimSpace(result)
}

func (s *Scraper) cleanCode(code string) string {
	// Remove HTML tags from code
	code = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(code, "")

	// Clean HTML entities
	code = strings.ReplaceAll(code, "&amp;", "&")
	code = strings.ReplaceAll(code, "&lt;", "<")
	code = strings.ReplaceAll(code, "&gt;", ">")
	code = strings.ReplaceAll(code, "&quot;", "\"")
	code = strings.ReplaceAll(code, "&#39;", "'")
	code = strings.ReplaceAll(code, "&nbsp;", " ")

	lines := strings.Split(code, "\n")
	var cleaned []string
	for _, line := range lines {
		cleaned = append(cleaned, strings.TrimRightFunc(line, func(r rune) bool {
			return r == ' ' || r == '\t'
		}))
	}
	return strings.Join(cleaned, "\n")
}

func (s *Scraper) SaveToFile(docs []types.DocSection, sessionDir, filename, format string) error {
	// Ensure session directory exists
	if err := s.ensureSessionDir(sessionDir); err != nil {
		return err
	}

	filepath := fmt.Sprintf("docs/%s/%s", sessionDir, filename)
	formatter := formatter.GetFormatter(format)
	data, err := formatter.FormatFull(docs)
	if err != nil {
		return fmt.Errorf("error formatting data: %w", err)
	}

	return s.writeFile(filepath, data)
}

func (s *Scraper) SaveSummaryToFile(docs []types.DocSection, sessionDir, filename string) error {
	// Ensure session directory exists
	if err := s.ensureSessionDir(sessionDir); err != nil {
		return err
	}

	filepath := fmt.Sprintf("docs/%s/%s", sessionDir, filename)
	generator := summary.New()
	summaryText := generator.GenerateReport(docs)

	return s.writeFile(filepath, []byte(summaryText))
}

func (s *Scraper) ensureSessionDir(sessionDir string) error {
	fullPath := fmt.Sprintf("docs/%s", sessionDir)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return os.MkdirAll(fullPath, 0755)
	}
	return nil
}

func (s *Scraper) writeFile(filename string, data []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}
