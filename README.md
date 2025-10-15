# PocketBase Documentation Scraper for LLMs

A specialized tool for scraping PocketBase documentation and optimizing it for Large Language Model (LLM) consumption. Built with Go and designed to generate AI-friendly documentation with comprehensive token analysis.

## 🎯 Purpose

This tool automatically scrapes the complete PocketBase documentation and generates:
- **LLM-optimized format** with structured metadata, token counting, and AI-friendly organization
- **Plain text backup** for general use
- **Comprehensive statistics** including token analysis, context window usage, and content insights

## 🚀 Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd pb-llm

# Run the scraper (generates both LLM and TXT formats automatically)
go run cmd/main.go

# Or build and run the binary
go build -o pb-tool cmd/main.go
./pb-tool
```

## 📁 Project Structure

```
pb-llm/
├── cmd/
│   └── main.go                  # Main entry point
├── internal/
│   ├── types/
│   │   └── types.go             # Data structures & token estimation
│   ├── scraper/
│   │   └── scraper.go           # Web scraping functionality
│   ├── formatter/
│   │   └── formatter.go         # Output formatting (LLM, TXT, etc.)
│   └── summary/
│       └── summary.go           # Statistics & summary generation
├── docs/                        # Generated documentation sessions
│   └── session_YYYY-MM-DD_HH-MM-SS.mmm/
│       ├── pocketbase_docs.llm.md    # LLM-optimized documentation
│       ├── pocketbase_docs.txt       # Plain text documentation
│       └── summary.txt               # Comprehensive statistics
├── go.mod
└── README.md
```

## 🤖 LLM Features

### Token Analysis
- **Accurate token estimation** for each section and the entire dataset
- **Context window usage calculation** (4K, 8K, 16K, etc.)
- **Compression ratio analysis** (clean vs raw content)
- **Per-category token breakdown**

### AI-Optimized Output
- **Structured metadata** for each documentation section
- **Clear section boundaries** with consistent formatting
- **Code examples** properly annotated for language detection
- **API parameters** clearly marked as required/optional
- **Response examples** with HTTP status codes

### Comprehensive Statistics
- **Largest/smallest sections** for content planning
- **Category-wise token distribution** for balanced training
- **Success/failure rates** for data quality assessment
- **API method coverage** analysis

## 📊 Example Output

### Summary Statistics
```
🤖 LLM USAGE STATISTICS
================================================================================
📊 Total Estimated Tokens: 135,048
🔧 LLM-Ready Tokens: 141,475
📈 Avg Tokens/Section: 2,774
📖 Context Window Usage: ~3454.0% (assuming 4k context)
🎯 Compression Ratio: 104.8% (clean vs raw)

📂 CONTENT BY CATEGORY (with Token Analysis)
================================================================================
   Hooks: 31 sections, 64,890 tokens (45.9%), avg 2,093 tokens/section
   Api: 10 sections, 51,715 tokens (36.6%), avg 5,171 tokens/section
   Auth: 1 sections, 7,188 tokens (5.1%), avg 7,188 tokens/section
```

### LLM-Optimized Documentation Format
```markdown
## SECTION 12: Record CRUD API

**METADATA:**
- URL: https://pocketbase.io/docs/api-records/
- Category: api
- Estimated Tokens: 20,480
- API Endpoint: `GET /api/collections/{collection}/records`

**PARAMETERS:**
- `collectionIdOrName` (String) [REQUIRED]: ID or name of the records' collection
- `page` (Number): The page offset of paginated list
- `perPage` (Number): Max returned records per page

**CODE EXAMPLES:**
```javascript
// JavaScript Example for Record CRUD API
const records = await pb.collection('posts').getList(1, 50, {
    filter: 'created >= "2022-01-01 00:00:00"'
});
```

## 🔧 Technical Details

### Token Estimation Algorithm
The tool uses a sophisticated estimation algorithm that considers:
- Word count (0.75 tokens per word average)
- Punctuation and symbols (0.5 tokens per character)
- Code blocks and technical content weighting
- Language-specific adjustments

### Content Processing
1. **Web scraping** with rate limiting and retry logic
2. **HTML parsing** with content extraction and cleaning  
3. **Parameter detection** from API documentation
4. **Code example extraction** with language detection
5. **Response example parsing** with status code identification

## 📝 Use Cases

### For LLM Training
- Use the `.llm.md` file as training data for PocketBase-focused models
- Token counts help with dataset balancing and context window planning
- Structured format ensures consistent model understanding

### For AI Applications
- Feed documentation into RAG (Retrieval-Augmented Generation) systems
- Use token estimates for chunk size optimization
- Category breakdown helps with targeted knowledge retrieval

### For Documentation Analysis
- Understand PocketBase API coverage and completeness
- Identify documentation gaps or inconsistencies  
- Track changes in documentation over time

## 🛠️ Advanced Usage

### Help Information
```bash
go run cmd/main.go -help
```

### Custom Build
```bash
# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o pb-tool-linux cmd/main.go
GOOS=windows GOARCH=amd64 go build -o pb-tool.exe cmd/main.go
GOOS=darwin GOARCH=amd64 go build -o pb-tool-mac cmd/main.go
```

## 📈 Output Files

| File | Purpose | Format |
|------|---------|--------|
| `pocketbase_docs.llm.md` | LLM training & AI applications | Structured Markdown with metadata |
| `pocketbase_docs.txt` | General use & backup | Plain text format |
| `summary.txt` | Statistics & analysis | Comprehensive report with token metrics |

## 🚦 Rate Limiting

The scraper includes built-in rate limiting:
- **1 second delay** between requests
- **Retry logic** with exponential backoff
- **Timeout handling** for reliable scraping

## 🎯 Goals

This tool is specifically designed for:
- **LLM developers** building PocketBase-aware AI systems
- **AI researchers** studying backend-as-a-service documentation
- **Developers** creating PocketBase-focused coding assistants
- **Data scientists** analyzing technical documentation patterns

## 📋 Requirements

- Go 1.19 or higher
- Internet connection for scraping
- ~50MB disk space for generated documentation

---

**Ready to enhance your LLM with comprehensive PocketBase knowledge!** 🚀