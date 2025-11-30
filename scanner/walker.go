package scanner

import (
	"os"
	"path/filepath"

	ignore "github.com/sabhiram/go-gitignore"
)

// IgnoredDirs are directories to skip during scanning
var IgnoredDirs = map[string]bool{
	".git":           true,
	"node_modules":   true,
	"vendor":         true,
	"Pods":           true,
	"build":          true,
	"DerivedData":    true,
	".idea":          true,
	".vscode":        true,
	"__pycache__":    true,
	".DS_Store":      true,
	"venv":           true,
	".venv":          true,
	".env":           true,
	".pytest_cache":  true,
	"dist":           true,
	".next":          true,
	".nuxt":          true,
	"target":         true,
	".grammar-build": true,
	"grammars":       true,
}

// LoadGitignore loads .gitignore from root if it exists
func LoadGitignore(root string) *ignore.GitIgnore {
	gitignorePath := filepath.Join(root, ".gitignore")

	if _, err := os.Stat(gitignorePath); err == nil {
		if gitignore, err := ignore.CompileIgnoreFile(gitignorePath); err == nil {
			return gitignore
		}
	}

	return nil
}

// ScanFiles walks the directory tree and returns all files
func ScanFiles(root string, gitignore *ignore.GitIgnore) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Skip if matched by common ignore patterns
		if info.IsDir() {
			if IgnoredDirs[info.Name()] {
				return filepath.SkipDir
			}
		} else {
			if IgnoredDirs[info.Name()] {
				return nil
			}
		}

		// Skip if matched by .gitignore
		if gitignore != nil && gitignore.MatchesPath(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories (we only want files in the output)
		if info.IsDir() {
			return nil
		}

		files = append(files, FileInfo{
			Path: relPath,
			Size: info.Size(),
			Ext:  filepath.Ext(path),
		})

		return nil
	})

	return files, err
}

// ScanForDeps walks the directory tree and analyzes files for dependencies
func ScanForDeps(root string, gitignore *ignore.GitIgnore, loader *GrammarLoader) ([]FileAnalysis, error) {
	var analyses []FileAnalysis

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(root, path)

		// Skip ignored dirs
		if info.IsDir() {
			if IgnoredDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if IgnoredDirs[info.Name()] {
			return nil
		}

		// Skip if matched by .gitignore
		if gitignore != nil && gitignore.MatchesPath(relPath) {
			return nil
		}

		// Only analyze supported languages
		if DetectLanguage(path) == "" {
			return nil
		}

		// Analyze file
		analysis, err := loader.AnalyzeFile(path)
		if err != nil || analysis == nil {
			return nil
		}

		// Use relative path in output
		analysis.Path = relPath
		analyses = append(analyses, *analysis)

		return nil
	})

	return analyses, err
}
