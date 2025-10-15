# PocketBase Docs Scraper for LLMs

```bash
# Clone the repository
git clone https://github.com/magooney-loon/pb-llm
cd pb-llm

# Run the scraper
go run cmd/main.go
```

```bash
PocketBase Documentation Scraper for LLM Usage
	=============================================

	DESCRIPTION:
	  Scrapes PocketBase documentation and automatically generates 4 variations:
	  • Full - Complete documentation with all extensions
	  • Go-only - Go extensions only (backend development)
	  • JS-only - JavaScript extensions only (frontend development)
	  • Core-only - Core PocketBase without any extensions

	  Each variation is generated in ultra-compact LLM-optimized and plain text formats.

	USAGE:
	  go run cmd/main.go [OPTIONS]

	OPTIONS:
	  -help
	        Show this help message

	OUTPUT FORMATS:
	  • .llm.md - Ultra-compact LLM format for maximum token efficiency
	  • .txt - Plain text format for general use

	FEATURES:
	  🤖 LLM-optimized output format
	  📊 Token counting and estimation
	  📈 Context window usage analysis
	  🔧 AI training dataset structure
	  📝 Comprehensive LLM usage statistics
	  📄 Plain text backup format
	  🎯 Automatic generation of all variations
	  📦 Pick exactly what you need

	OUTPUT (4 variations × 2 formats = 8 documentation files):
	  • pocketbase_docs_full.llm.md/.txt - Complete documentation
	  • pocketbase_docs_go.llm.md/.txt - Go extensions only
	  • pocketbase_docs_js.llm.md/.txt - JavaScript extensions only
	  • pocketbase_docs_core.llm.md/.txt - Core PocketBase only
	  • summary_*.txt - Individual statistics for each variation

	EXAMPLE:
	  go run cmd/main.go                      # Generates all 4 variations

	All files saved in timestamped docs/session_YYYY-MM-DD_HH-MM-SS.mmm/ directory
```
