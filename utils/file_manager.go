// --- START OF FILE utils/file_manager.go ---
package utils

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// --- Indentation Logic ---

type IndentTypeStr string

const (
	IndentSpace IndentTypeStr = "space"
	IndentTab   IndentTypeStr = "tab"
	IndentMixed IndentTypeStr = "mixed"
)

type IndentType struct {
	Type     IndentTypeStr
	Size     int
	MostUsed *IndentType // For mixed
}

func detectLineIndent(line string) (int, int) {
	tabs := 0
	spaces := 0
	for _, char := range line {
		if char == '\t' {
			tabs++
		} else {
			break
		}
	}
	for _, char := range line[tabs:] {
		if char == ' ' {
			spaces++
		} else {
			break
		}
	}
	return tabs, spaces
}

func DetectIndentType(code string) *IndentType {
	lines := strings.Split(code, "\n")
	spaceDiffCounts := make(map[int]int)
	tabIndents := 0
	spaceIndents := 0
	prevSpaces := 0
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		tabs, spaces := detectLineIndent(line)
		if tabs == 0 && spaces == 0 {
			continue
		}
		if tabs > 0 {
			tabIndents++
		} else {
			spaceIndents++
			diff := abs(spaces - prevSpaces)
			if diff > 1 {
				spaceDiffCounts[diff]++
			}
			prevSpaces = spaces
		}
	}

	if tabIndents > 0 && spaceIndents > 0 {
		mostUsed := &IndentType{Type: IndentSpace, Size: 4}
		if tabIndents > spaceIndents {
			mostUsed = &IndentType{Type: IndentTab, Size: 1}
		} else {
			mostUsed = getMostCommonSpace(spaceDiffCounts)
		}
		return &IndentType{Type: IndentMixed, Size: 1, MostUsed: mostUsed}
	} else if tabIndents > 0 {
		return &IndentType{Type: IndentTab, Size: 1}
	} else if len(spaceDiffCounts) > 0 {
		return getMostCommonSpace(spaceDiffCounts)
	}
	
	return nil // Default or none
}

func getMostCommonSpace(counts map[int]int) *IndentType {
	maxVal := 0
	size := 4
	for k, v := range counts {
		if v > maxVal {
			maxVal = v
			size = k
		}
	}
	return &IndentType{Type: IndentSpace, Size: size}
}

func MatchIndent(code, template string) string {
	it := DetectIndentType(template)
	if it == nil {
		return code
	}
	if it.Type == IndentMixed && it.MostUsed != nil {
		it = it.MostUsed
	}
	
	lines := strings.Split(code, "\n")
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, line)
			continue
		}
		tabs, spaces := detectLineIndent(line)
		levels := 0
		if tabs > 0 {
			levels = tabs
		} else {
			levels = spaces / 4 // Assuming input is normalized to 4 spaces or we calc ratio
			if it.Size > 0 {
				levels = spaces / 4 
			}
		}
		
		indent := ""
		if it.Type == IndentTab {
			indent = strings.Repeat("\t", levels)
		} else {
			indent = strings.Repeat(" ", levels*it.Size)
		}
		result = append(result, indent+strings.TrimLeft(line, " \t"))
	}
	return strings.Join(result, "\n")
}

func matchIndentByFirstLine(newStr, refLine string) string {
	_, targetSpaces := detectLineIndent(refLine)
	lines := strings.Split(newStr, "\n")
	if len(lines) == 0 {
		return newStr
	}
	_, currentSpaces := detectLineIndent(lines[0])
	diff := targetSpaces - currentSpaces
	
	var res []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			res = append(res, line)
			continue
		}
		_, spaces := detectLineIndent(line)
		newIndent := spaces + diff
		if newIndent < 0 {
			newIndent = 0
		}
		res = append(res, strings.Repeat(" ", newIndent)+strings.TrimLeft(line, " \t"))
	}
	return strings.Join(res, "\n")
}

// --- StrReplaceManager ---

type StrReplaceManager struct {
	History          map[string][]string // Path -> History
	IgnoreIndentation bool
	ExpandTabs        bool
	mu               sync.Mutex
}

func NewStrReplaceManager(ignoreIndent, expandTabs bool) *StrReplaceManager {
	return &StrReplaceManager{
		History:           make(map[string][]string),
		IgnoreIndentation: ignoreIndent,
		ExpandTabs:        expandTabs,
	}
}

func (m *StrReplaceManager) ReadFile(pathStr string) StrReplaceResponse {
	content, err := os.ReadFile(pathStr)
	if err != nil {
		return StrReplaceResponse{Success: false, FileContent: err.Error()}
	}
	return StrReplaceResponse{Success: true, FileContent: string(content)}
}

func (m *StrReplaceManager) WriteFile(pathStr, content string) StrReplaceResponse {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Backup for undo
	if current, err := os.ReadFile(pathStr); err == nil {
		m.History[pathStr] = append(m.History[pathStr], string(current))
	}

	err := os.WriteFile(pathStr, []byte(content), 0644)
	if err != nil {
		return StrReplaceResponse{Success: false, FileContent: err.Error()}
	}
	return StrReplaceResponse{Success: true, FileContent: content}
}

func (m *StrReplaceManager) StrReplace(pathStr, oldStr, newStr string) StrReplaceResponse {
	m.mu.Lock()
	defer m.mu.Unlock()

	contentBytes, err := os.ReadFile(pathStr)
	if err != nil {
		return StrReplaceResponse{Success: false, FileContent: fmt.Sprintf("Read error: %v", err)}
	}
	content := string(contentBytes)
	
	if m.ExpandTabs {
		content = strings.ReplaceAll(content, "\t", "    ")
		oldStr = strings.ReplaceAll(oldStr, "\t", "    ")
		newStr = strings.ReplaceAll(newStr, "\t", "    ")
	}

	// Simple case
	if !m.IgnoreIndentation {
		if strings.Count(content, oldStr) == 0 {
			return StrReplaceResponse{Success: false, FileContent: "old_str not found verbatim."}
		}
		if strings.Count(content, oldStr) > 1 {
			return StrReplaceResponse{Success: false, FileContent: "Multiple occurrences of old_str found."}
		}
		
		newContent := strings.Replace(content, oldStr, newStr, 1)
		
		// History
		m.History[pathStr] = append(m.History[pathStr], content)
		
		if err := os.WriteFile(pathStr, []byte(newContent), 0644); err != nil {
			return StrReplaceResponse{Success: false, FileContent: err.Error()}
		}
		return StrReplaceResponse{Success: true, FileContent: makeSnippet(newContent, newStr)}
	}

	// Complex Indentation Ignoring Logic
	lines := strings.Split(content, "\n")
	oldLines := strings.Split(oldStr, "\n")
	strippedLines := make([]string, len(lines))
	for i, l := range lines {
		strippedLines[i] = strings.TrimSpace(l)
	}
	strippedOld := make([]string, len(oldLines))
	for i, l := range oldLines {
		strippedOld[i] = strings.TrimSpace(l)
	}

	matchIndex := -1
	for i := 0; i <= len(lines)-len(oldLines); i++ {
		match := true
		for j := 0; j < len(oldLines); j++ {
			if strippedLines[i+j] != strippedOld[j] {
				// Handle partial match on last line if needed, but strict line matching is safer for now
				if j == len(oldLines)-1 && strings.HasPrefix(strippedLines[i+j], strippedOld[j]) {
					// Soft match for end
				} else {
					match = false
					break
				}
			}
		}
		if match {
			if matchIndex != -1 {
				return StrReplaceResponse{Success: false, FileContent: "Multiple fuzzy matches found."}
			}
			matchIndex = i
		}
	}

	if matchIndex == -1 {
		return StrReplaceResponse{Success: false, FileContent: "No match found."}
	}

	// Reconstruct
	// Match indentation of first line
	indentedNewStr := matchIndentByFirstLine(newStr, lines[matchIndex])
	
	newContentLines := append(lines[:matchIndex], strings.Split(indentedNewStr, "\n")...)
	newContentLines = append(newContentLines, lines[matchIndex+len(oldLines):]...)
	finalContent := strings.Join(newContentLines, "\n")

	m.History[pathStr] = append(m.History[pathStr], content)
	os.WriteFile(pathStr, []byte(finalContent), 0644)

	return StrReplaceResponse{Success: true, FileContent: makeSnippet(finalContent, indentedNewStr)}
}

func (m *StrReplaceManager) Undo(pathStr string) StrReplaceResponse {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	hist, ok := m.History[pathStr]
	if !ok || len(hist) == 0 {
		return StrReplaceResponse{Success: false, FileContent: "No history found."}
	}
	
	prev := hist[len(hist)-1]
	m.History[pathStr] = hist[:len(hist)-1]
	
	if err := os.WriteFile(pathStr, []byte(prev), 0644); err != nil {
		return StrReplaceResponse{Success: false, FileContent: err.Error()}
	}
	return StrReplaceResponse{Success: true, FileContent: "Undo successful"}
}

// Helper
func makeSnippet(fullContent, changeBlock string) string {
	// Simplified snippet generation
	if len(fullContent) > 500 {
		return "File edited. (Content truncated for brevity)"
	}
	return fullContent
}

func abs(x int) int {
	if x < 0 { return -x }
	return x
}