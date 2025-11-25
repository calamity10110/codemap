package scanner

// FileInfo represents a single file in the codebase.
type FileInfo struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Ext  string `json:"ext"`
}

// Project represents the root of the codebase for tree/skyline mode.
type Project struct {
	Root    string     `json:"root"`
	Mode    string     `json:"mode"`
	Animate bool       `json:"animate"`
	Files   []FileInfo `json:"files"`
}

// FileAnalysis holds extracted info about a single file for deps mode.
type FileAnalysis struct {
	Path      string   `json:"path"`
	Language  string   `json:"language"`
	Functions []string `json:"functions"`
	Imports   []string `json:"imports"`
}

// DepsProject is the JSON output for --deps mode.
type DepsProject struct {
	Root         string              `json:"root"`
	Mode         string              `json:"mode"`
	Files        []FileAnalysis      `json:"files"`
	ExternalDeps map[string][]string `json:"external_deps"`
}
