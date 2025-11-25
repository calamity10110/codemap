package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"codemap/render"
	"codemap/scanner"

	ignore "github.com/sabhiram/go-gitignore"
)

func main() {
	skylineMode := flag.Bool("skyline", false, "Enable skyline visualization mode")
	animateMode := flag.Bool("animate", false, "Enable animation (use with --skyline)")
	depsMode := flag.Bool("deps", false, "Enable dependency graph mode (function/import analysis)")
	jsonMode := flag.Bool("json", false, "Output JSON (for Python renderer compatibility)")
	helpMode := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *helpMode {
		fmt.Println("codemap - Generate a brain map of your codebase for LLM context")
		fmt.Println()
		fmt.Println("Usage: codemap [options] [path]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  --help      Show this help message")
		fmt.Println("  --skyline   City skyline visualization")
		fmt.Println("  --animate   Animated skyline (use with --skyline)")
		fmt.Println("  --deps      Dependency flow map (functions & imports)")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  codemap .                    # Basic tree view")
		fmt.Println("  codemap --skyline .          # Skyline visualization")
		fmt.Println("  codemap --skyline --animate  # Animated skyline")
		fmt.Println("  codemap --deps /path/to/proj # Dependency flow map")
		os.Exit(0)
	}

	root := flag.Arg(0)
	if root == "" {
		root = "."
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting absolute path: %v\n", err)
		os.Exit(1)
	}

	// Load .gitignore if it exists
	gitignore := scanner.LoadGitignore(root)

	// Handle --deps mode separately
	if *depsMode {
		runDepsMode(absRoot, root, gitignore, *jsonMode)
		return
	}

	mode := "tree"
	if *skylineMode {
		mode = "skyline"
	}

	// Scan files
	files, err := scanner.ScanFiles(root, gitignore)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking tree: %v\n", err)
		os.Exit(1)
	}

	project := scanner.Project{
		Root:    absRoot,
		Mode:    mode,
		Animate: *animateMode,
		Files:   files,
	}

	// Render or output JSON
	if *jsonMode {
		json.NewEncoder(os.Stdout).Encode(project)
	} else if *skylineMode {
		render.Skyline(project, *animateMode)
	} else {
		render.Tree(project)
	}
}

func runDepsMode(absRoot, root string, gitignore *ignore.GitIgnore, jsonMode bool) {
	loader := scanner.NewGrammarLoader()

	analyses, err := scanner.ScanForDeps(root, gitignore, loader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning for deps: %v\n", err)
		os.Exit(1)
	}

	depsProject := scanner.DepsProject{
		Root:         absRoot,
		Mode:         "deps",
		Files:        analyses,
		ExternalDeps: scanner.ReadExternalDeps(absRoot),
	}

	// Render or output JSON
	if jsonMode {
		json.NewEncoder(os.Stdout).Encode(depsProject)
	} else {
		render.Depgraph(depsProject)
	}
}
