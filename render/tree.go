package render

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"codemap/scanner"
)

// treeNode represents a node in the file tree
type treeNode struct {
	name     string
	isFile   bool
	file     *scanner.FileInfo
	children map[string]*treeNode
}

// getTopLargeFiles returns paths of top 5 largest source code files
func getTopLargeFiles(files []scanner.FileInfo) map[string]bool {
	// Filter out assets and binaries (no extension = likely binary)
	var sourceFiles []scanner.FileInfo
	for _, f := range files {
		ext := strings.ToLower(f.Ext)
		// Skip if no extension (likely binary) or if it's an asset
		if ext == "" || IsAssetExtension(ext) {
			continue
		}
		sourceFiles = append(sourceFiles, f)
	}

	// Sort by size descending
	sort.Slice(sourceFiles, func(i, j int) bool {
		return sourceFiles[i].Size > sourceFiles[j].Size
	})

	// Return top 5 as set
	result := make(map[string]bool)
	for i := 0; i < len(sourceFiles) && i < 5; i++ {
		result[sourceFiles[i].Path] = true
	}
	return result
}

// getDirStats recursively calculates file count and total size
func getDirStats(node *treeNode) (int, int64) {
	if node.isFile {
		return 1, node.file.Size
	}
	count := 0
	var size int64 = 0
	for _, child := range node.children {
		c, s := getDirStats(child)
		count += c
		size += s
	}
	return count, size
}

// buildTreeStructure builds a nested tree from flat file list
func buildTreeStructure(files []scanner.FileInfo) *treeNode {
	root := &treeNode{children: make(map[string]*treeNode)}

	for _, f := range files {
		parts := strings.Split(f.Path, string(os.PathSeparator))
		current := root
		for i, part := range parts {
			if i == len(parts)-1 {
				// File
				fileCopy := f
				current.children[part] = &treeNode{
					name:   part,
					isFile: true,
					file:   &fileCopy,
				}
			} else {
				// Directory
				if current.children[part] == nil {
					current.children[part] = &treeNode{
						name:     part,
						children: make(map[string]*treeNode),
					}
				}
				current = current.children[part]
			}
		}
	}
	return root
}

// formatSize converts bytes to human readable format
func formatSize(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	fsize := float64(size)
	for _, unit := range units[:len(units)-1] {
		if fsize < 1024 {
			return fmt.Sprintf("%.1f%s", fsize, unit)
		}
		fsize /= 1024
	}
	return fmt.Sprintf("%.1f%s", fsize, units[len(units)-1])
}

// Tree renders the file tree to stdout
func Tree(project scanner.Project) {
	files := project.Files
	projectName := filepath.Base(project.Root)

	// Calculate stats
	totalFiles := len(files)
	var totalSize int64 = 0
	extCount := make(map[string]int)
	for _, f := range files {
		totalSize += f.Size
		if f.Ext != "" {
			extCount[f.Ext]++
		}
	}

	// Get top extensions
	type extEntry struct {
		ext   string
		count int
	}
	var exts []extEntry
	for ext, count := range extCount {
		exts = append(exts, extEntry{ext, count})
	}
	sort.Slice(exts, func(i, j int) bool {
		return exts[i].count > exts[j].count
	})
	if len(exts) > 5 {
		exts = exts[:5]
	}

	// Get top large files
	topLarge := getTopLargeFiles(files)

	// Print header (match Python rich panel exactly - title in top border)
	innerWidth := 64

	// Title in top border line (like rich panel)
	titleLine := fmt.Sprintf(" %s ", projectName)
	padding := innerWidth - len(titleLine)
	leftPad := padding / 2
	rightPad := padding - leftPad
	fmt.Printf("╭%s%s%s╮\n", strings.Repeat("─", leftPad), titleLine, strings.Repeat("─", rightPad))

	// Stats line
	statsLine := fmt.Sprintf("Files: %d | Size: %s", totalFiles, formatSize(totalSize))
	fmt.Printf("│ %-*s │\n", innerWidth-2, statsLine)

	// Extensions line
	if len(exts) > 0 {
		extParts := make([]string, len(exts))
		for i, e := range exts {
			extParts[i] = fmt.Sprintf("%s (%d)", e.ext, e.count)
		}
		extLine := "Top Extensions: " + strings.Join(extParts, ", ")
		fmt.Printf("│ %-*s │\n", innerWidth-2, extLine)
	}

	fmt.Printf("╰%s╯\n", strings.Repeat("─", innerWidth))

	// Build and render tree
	root := buildTreeStructure(files)
	fmt.Printf("%s%s%s\n", Bold, projectName, Reset)
	printTreeNode(root, "", true, topLarge)
}

// printTreeNode recursively prints tree nodes
func printTreeNode(node *treeNode, prefix string, isLast bool, topLarge map[string]bool) {
	// Separate dirs and files
	var dirs, fileNodes []*treeNode
	for _, child := range node.children {
		if child.isFile {
			fileNodes = append(fileNodes, child)
		} else {
			dirs = append(dirs, child)
		}
	}

	// Sort
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].name < dirs[j].name })
	sort.Slice(fileNodes, func(i, j int) bool { return fileNodes[i].name < fileNodes[j].name })

	// Print directories first
	for i, dir := range dirs {
		isLastDir := i == len(dirs)-1 && len(fileNodes) == 0

		// Flatten single-child directories
		mergedName := dir.name
		current := dir
		for len(current.children) == 1 {
			var onlyChild *treeNode
			for _, c := range current.children {
				onlyChild = c
			}
			if onlyChild.isFile {
				break
			}
			mergedName = mergedName + "/" + onlyChild.name
			current = onlyChild
		}

		// Get stats
		fileCount, totalSize := getDirStats(current)

		// Check for homogeneous extensions
		var commonExt string
		immediateFiles := make([]*scanner.FileInfo, 0)
		for _, c := range current.children {
			if c.isFile {
				immediateFiles = append(immediateFiles, c.file)
			}
		}
		if len(immediateFiles) > 1 {
			extSet := make(map[string]bool)
			for _, f := range immediateFiles {
				extSet[f.Ext] = true
			}
			if len(extSet) == 1 {
				for ext := range extSet {
					commonExt = ext
				}
			}
		}

		// Format stats
		var statsParts []string
		if fileCount == 1 {
			statsParts = append(statsParts, formatSize(totalSize))
		} else {
			statsParts = append(statsParts, fmt.Sprintf("%d files", fileCount))
			statsParts = append(statsParts, formatSize(totalSize))
		}
		if commonExt != "" {
			statsParts = append(statsParts, fmt.Sprintf("all %s", commonExt))
		}

		connector := "├── "
		if isLastDir {
			connector = "└── "
		}

		fmt.Printf("%s%s%s  %s/%s %s(%s)%s\n",
			prefix, connector, BoldBlue, mergedName, Reset, Dim, strings.Join(statsParts, ", "), Reset)

		newPrefix := prefix + "│   "
		if isLastDir {
			newPrefix = prefix + "    "
		}
		printTreeNode(current, newPrefix, isLastDir, topLarge)
	}

	// Print files as a grid (multi-column layout like Python)
	if len(fileNodes) > 0 {
		connector := "└── "
		termWidth := GetTerminalWidth()
		availableWidth := termWidth - len(prefix) - len(connector)
		if availableWidth < 40 {
			availableWidth = 40
		}

		// Check if all files have the same extension (strip if so, like Python)
		stripExt := ""
		if len(fileNodes) > 1 {
			extSet := make(map[string]bool)
			for _, f := range fileNodes {
				extSet[f.file.Ext] = true
			}
			if len(extSet) == 1 {
				for ext := range extSet {
					stripExt = ext
				}
			}
		}

		// Build file entries with colors
		type fileEntry struct {
			display string
			colored string
			width   int
		}
		var entries []fileEntry
		for _, f := range fileNodes {
			color := GetFileColor(f.file.Ext)
			displayName := f.name
			// Strip extension if all files have same extension
			if stripExt != "" {
				displayName = strings.TrimSuffix(displayName, stripExt)
				if displayName == "" {
					displayName = f.name // Keep original if stripping leaves empty
				}
			}

			// Star for top large files
			starPrefix := ""
			if topLarge[f.file.Path] {
				starPrefix = "⭐️ "
				color = Bold + color
			}

			display := starPrefix + displayName
			colored := fmt.Sprintf("%s%s%s%s", color, starPrefix, displayName, Reset)
			// Width calculation: star emoji is 2 chars wide visually
			width := len(displayName)
			if starPrefix != "" {
				width += 3 // "⭐️ " is about 3 chars wide
			}
			entries = append(entries, fileEntry{display, colored, width})
		}

		// Calculate columns - find max width and fit columns
		maxWidth := 0
		for _, e := range entries {
			if e.width > maxWidth {
				maxWidth = e.width
			}
		}
		colWidth := maxWidth + 1
		numCols := availableWidth / colWidth
		if numCols < 1 {
			numCols = 1
		}
		if numCols > len(entries) {
			numCols = len(entries)
		}

		// Calculate number of rows
		numRows := (len(entries) + numCols - 1) / numCols

		// Print in column-major order (like Python)
		for row := 0; row < numRows; row++ {
			if row == 0 {
				fmt.Printf("%s%s", prefix, connector)
			} else {
				fmt.Printf("%s    ", prefix)
			}
			for col := 0; col < numCols; col++ {
				idx := col*numRows + row
				if idx < len(entries) {
					e := entries[idx]
					// Pad to column width
					padding := colWidth - e.width
					if padding < 0 {
						padding = 0
					}
					fmt.Printf("%s%s", e.colored, strings.Repeat(" ", padding))
				}
			}
			fmt.Println()
		}
	}
}
