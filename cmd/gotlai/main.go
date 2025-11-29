// Command gotlai translates HTML files using AI.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZaguanLabs/gotlai"
	"github.com/ZaguanLabs/gotlai/cache"
	"github.com/ZaguanLabs/gotlai/processor"
	"github.com/ZaguanLabs/gotlai/provider"
)

// Build-time variables (can be overridden with ldflags)
var (
	version   = gotlai.Version
	commit    = gotlai.GitCommit
	buildDate = gotlai.BuildDate
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("gotlai", flag.ContinueOnError)
	fs.SetOutput(stderr)

	// Flags
	targetLang := fs.String("lang", "", "Target language code (e.g., es_ES, ja_JP)")
	sourceLang := fs.String("source", "en", "Source language code")
	output := fs.String("output", "", "Output file (default: stdout)")
	outputShort := fs.String("o", "", "Output file (short for --output)")
	apiKey := fs.String("api-key", "", "OpenAI API key (default: OPENAI_API_KEY env)")
	model := fs.String("model", "gpt-4o-mini", "OpenAI model to use")
	contextStr := fs.String("context", "", "Translation context (e.g., 'E-commerce website')")
	exclude := fs.String("exclude", "", "Comma-separated terms to never translate")
	cacheTTL := fs.Int("cache-ttl", 3600, "Cache TTL in seconds (0 to disable)")
	showVersion := fs.Bool("version", false, "Show version")
	quiet := fs.Bool("quiet", false, "Suppress progress output")
	dryRun := fs.Bool("dry-run", false, "Show what would be translated without calling API")
	jsonOutput := fs.Bool("json", false, "Output result as JSON")
	diffFile := fs.String("diff", "", "Compare with previous version and show changes")
	updateMode := fs.Bool("update", false, "Only translate new/changed content (requires --diff)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *showVersion {
		fmt.Fprintf(stdout, "%s %s\n", gotlai.Name, version)
		if commit != "unknown" && commit != "" {
			fmt.Fprintf(stdout, "  commit:  %s\n", commit)
		}
		if buildDate != "unknown" && buildDate != "" {
			fmt.Fprintf(stdout, "  built:   %s\n", buildDate)
		}
		return nil
	}

	// Handle -o alias for --output
	if *outputShort != "" && *output == "" {
		*output = *outputShort
	}

	// Validate required flags
	if *targetLang == "" {
		fs.Usage()
		return fmt.Errorf("--lang is required")
	}

	// Get input
	var input string
	var inputName string

	if fs.NArg() == 0 {
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		input = string(data)
		inputName = "stdin"
	} else {
		// Read from file - user-provided path is intentional for CLI
		inputPath := fs.Arg(0)
		data, err := os.ReadFile(inputPath) // #nosec G304 - CLI tool reads user-specified files
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}
		input = string(data)
		inputName = filepath.Base(inputPath)
	}

	// Handle diff mode
	if *diffFile != "" {
		return runDiff(input, *diffFile, inputName, *targetLang, stdout, stderr, *jsonOutput, *updateMode)
	}

	// Handle dry-run mode
	if *dryRun {
		return runDryRun(input, inputName, *targetLang, stdout, stderr, *jsonOutput)
	}

	// Get API key
	key := *apiKey
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	if key == "" {
		return fmt.Errorf("OpenAI API key required (--api-key or OPENAI_API_KEY env)")
	}

	// Create provider
	p := provider.NewOpenAIProvider(provider.OpenAIConfig{
		APIKey: key,
		Model:  *model,
	})

	// Wrap with retry
	retryable := gotlai.NewRetryableProvider(p, gotlai.DefaultRetryConfig())

	// Build options
	opts := []gotlai.TranslatorOption{
		gotlai.WithSourceLang(*sourceLang),
		gotlai.WithProcessor(processor.NewHTMLProcessor()),
	}

	if *cacheTTL > 0 {
		opts = append(opts, gotlai.WithCache(cache.NewInMemoryCache(*cacheTTL)))
	}

	if *contextStr != "" {
		opts = append(opts, gotlai.WithContext(*contextStr))
	}

	if *exclude != "" {
		terms := strings.Split(*exclude, ",")
		for i := range terms {
			terms[i] = strings.TrimSpace(terms[i])
		}
		opts = append(opts, gotlai.WithExcludedTerms(terms))
	}

	// Create translator
	translator := gotlai.NewTranslator(*targetLang, retryable, opts...)

	// Translate
	if !*quiet {
		fmt.Fprintf(stderr, "Translating %s to %s...\n", inputName, *targetLang)
	}

	start := time.Now()
	result, err := translator.ProcessHTML(context.Background(), input)
	if err != nil {
		return fmt.Errorf("translation failed: %w", err)
	}
	elapsed := time.Since(start)

	// Output
	var out io.Writer = stdout
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		out = f
	}

	if *jsonOutput {
		return outputJSON(out, result, elapsed)
	}

	fmt.Fprint(out, result.Content)

	// Stats
	if !*quiet {
		fmt.Fprintf(stderr, "\nDone in %v\n", elapsed.Round(time.Millisecond))
		fmt.Fprintf(stderr, "  Nodes found:  %d\n", result.TotalNodes)
		fmt.Fprintf(stderr, "  Translated:   %d\n", result.TranslatedCount)
		fmt.Fprintf(stderr, "  From cache:   %d\n", result.CachedCount)
	}

	return nil
}

// runDryRun shows what would be translated without calling the API.
func runDryRun(input, inputName, targetLang string, stdout, stderr io.Writer, jsonOut bool) error {
	proc := processor.NewHTMLProcessor()
	_, nodes, err := proc.Extract(input)
	if err != nil {
		return fmt.Errorf("extracting text: %w", err)
	}

	if jsonOut {
		type dryRunOutput struct {
			InputFile  string   `json:"input_file"`
			TargetLang string   `json:"target_lang"`
			NodeCount  int      `json:"node_count"`
			Texts      []string `json:"texts"`
		}

		texts := make([]string, len(nodes))
		for i, n := range nodes {
			texts[i] = n.Text
		}

		out := dryRunOutput{
			InputFile:  inputName,
			TargetLang: targetLang,
			NodeCount:  len(nodes),
			Texts:      texts,
		}

		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	fmt.Fprintf(stdout, "Dry run: %s -> %s\n", inputName, targetLang)
	fmt.Fprintf(stdout, "Found %d translatable text nodes:\n\n", len(nodes))

	for i, node := range nodes {
		text := node.Text
		if len(text) > 60 {
			text = text[:57] + "..."
		}
		fmt.Fprintf(stdout, "%3d. %q\n", i+1, text)
		if node.Context != "" {
			fmt.Fprintf(stdout, "     Context: %s\n", node.Context)
		}
	}

	return nil
}

// runDiff compares new content with a previous version and shows what changed.
func runDiff(newContent, oldPath, inputName, targetLang string, stdout, stderr io.Writer, jsonOut, updateMode bool) error {
	// Read old file
	oldData, err := os.ReadFile(oldPath) // #nosec G304 - CLI tool reads user-specified files
	if err != nil {
		return fmt.Errorf("reading previous version: %w", err)
	}

	proc := processor.NewHTMLProcessor()

	// Extract nodes from both versions
	_, oldNodes, err := proc.Extract(string(oldData))
	if err != nil {
		return fmt.Errorf("parsing previous version: %w", err)
	}

	_, newNodes, err := proc.Extract(newContent)
	if err != nil {
		return fmt.Errorf("parsing new version: %w", err)
	}

	// Compute diff
	diff := gotlai.DiffContentWithContext(oldNodes, newNodes)
	stats := diff.Stats()

	if jsonOut {
		type diffOutput struct {
			InputFile    string `json:"input_file"`
			PreviousFile string `json:"previous_file"`
			TargetLang   string `json:"target_lang"`
			Stats        struct {
				Added     int `json:"added"`
				Removed   int `json:"removed"`
				Modified  int `json:"modified"`
				Unchanged int `json:"unchanged"`
			} `json:"stats"`
			NeedsTranslation []string `json:"needs_translation"`
			Added            []string `json:"added,omitempty"`
			Removed          []string `json:"removed,omitempty"`
			Modified         []struct {
				Old string `json:"old"`
				New string `json:"new"`
			} `json:"modified,omitempty"`
		}

		out := diffOutput{
			InputFile:    inputName,
			PreviousFile: filepath.Base(oldPath),
			TargetLang:   targetLang,
		}
		out.Stats.Added = stats.Added
		out.Stats.Removed = stats.Removed
		out.Stats.Modified = stats.Modified
		out.Stats.Unchanged = stats.Unchanged

		for _, n := range diff.NeedsTranslation() {
			out.NeedsTranslation = append(out.NeedsTranslation, n.Text)
		}
		for _, n := range diff.Added {
			out.Added = append(out.Added, n.Text)
		}
		for _, n := range diff.Removed {
			out.Removed = append(out.Removed, n.Text)
		}
		for _, m := range diff.Modified {
			out.Modified = append(out.Modified, struct {
				Old string `json:"old"`
				New string `json:"new"`
			}{Old: m.Old.Text, New: m.New.Text})
		}

		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	// Text output
	fmt.Fprintf(stdout, "Diff: %s vs %s\n", inputName, filepath.Base(oldPath))
	fmt.Fprintf(stdout, "Target language: %s\n\n", targetLang)

	fmt.Fprintf(stdout, "Summary:\n")
	fmt.Fprintf(stdout, "  Unchanged: %d\n", stats.Unchanged)
	fmt.Fprintf(stdout, "  Added:     %d\n", stats.Added)
	fmt.Fprintf(stdout, "  Removed:   %d\n", stats.Removed)
	fmt.Fprintf(stdout, "  Modified:  %d\n", stats.Modified)
	fmt.Fprintf(stdout, "\n")

	if !diff.HasChanges() {
		fmt.Fprintf(stdout, "No changes detected. All translations are up to date.\n")
		return nil
	}

	needsTranslation := diff.NeedsTranslation()
	fmt.Fprintf(stdout, "Needs translation: %d strings\n\n", len(needsTranslation))

	if len(diff.Added) > 0 {
		fmt.Fprintf(stdout, "Added:\n")
		for _, n := range diff.Added {
			text := n.Text
			if len(text) > 50 {
				text = text[:47] + "..."
			}
			fmt.Fprintf(stdout, "  + %q\n", text)
		}
		fmt.Fprintf(stdout, "\n")
	}

	if len(diff.Modified) > 0 {
		fmt.Fprintf(stdout, "Modified:\n")
		for _, m := range diff.Modified {
			oldText := m.Old.Text
			newText := m.New.Text
			if len(oldText) > 30 {
				oldText = oldText[:27] + "..."
			}
			if len(newText) > 30 {
				newText = newText[:27] + "..."
			}
			fmt.Fprintf(stdout, "  ~ %q -> %q\n", oldText, newText)
		}
		fmt.Fprintf(stdout, "\n")
	}

	if len(diff.Removed) > 0 {
		fmt.Fprintf(stdout, "Removed:\n")
		for _, n := range diff.Removed {
			text := n.Text
			if len(text) > 50 {
				text = text[:47] + "..."
			}
			fmt.Fprintf(stdout, "  - %q\n", text)
		}
		fmt.Fprintf(stdout, "\n")
	}

	if updateMode {
		fmt.Fprintf(stdout, "Update mode: Only the %d new/modified strings would be translated.\n", len(needsTranslation))
		fmt.Fprintf(stdout, "Run without --diff to perform the translation.\n")
	}

	return nil
}

// JSONOutput represents the JSON output format.
type JSONOutput struct {
	Content         string `json:"content"`
	TotalNodes      int    `json:"total_nodes"`
	TranslatedCount int    `json:"translated_count"`
	CachedCount     int    `json:"cached_count"`
	ElapsedMs       int64  `json:"elapsed_ms"`
}

// outputJSON writes the result as JSON.
func outputJSON(w io.Writer, result *gotlai.ProcessedContent, elapsed time.Duration) error {
	out := JSONOutput{
		Content:         result.Content,
		TotalNodes:      result.TotalNodes,
		TranslatedCount: result.TranslatedCount,
		CachedCount:     result.CachedCount,
		ElapsedMs:       elapsed.Milliseconds(),
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
