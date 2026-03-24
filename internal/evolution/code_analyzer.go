package evolution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// CodeAnalyzer analyzes code and provides improvement suggestions
type CodeAnalyzer struct {
	accessor *CodeAccessor
	llm      LLMAnalyzer
}

// LLMAnalyzer is the interface for LLM-based code analysis
type LLMAnalyzer interface {
	AnalyzeCode(content string) ([]Suggestion, error)
}

// Suggestion represents a code improvement suggestion
type Suggestion struct {
	Type       string  `json:"type"`
	Line       int     `json:"line"`
	Message    string  `json:"message"`
	Suggestion string  `json:"suggestion"`
	Confidence float64 `json:"confidence"`
}

// NewCodeAnalyzer creates a new code analyzer
func NewCodeAnalyzer(accessor *CodeAccessor) *CodeAnalyzer {
	return &CodeAnalyzer{
		accessor: accessor,
	}
}

// SetLLM sets the LLM analyzer
func (my *CodeAnalyzer) SetLLM(llm LLMAnalyzer) {
	my.llm = llm
}

// AnalyzeFile analyzes a single file
func (my *CodeAnalyzer) AnalyzeFile(filePath string) ([]Suggestion, error) {
	content, err := my.accessor.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Use LLM if available
	if my.llm != nil {
		return my.llm.AnalyzeCode(string(content))
	}

	// Basic analysis
	return my.basicAnalysis(filePath, content)
}

// basicAnalysis performs basic static analysis
func (my *CodeAnalyzer) basicAnalysis(filePath string, content []byte) ([]Suggestion, error) {
	var suggestions []Suggestion

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return suggestions, nil // Return empty if parse fails
	}

	// Check for functions without comments
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Doc == nil {
				suggestions = append(suggestions, Suggestion{
					Type:       "documentation",
					Line:       fset.Position(x.Pos()).Line,
					Message:    fmt.Sprintf("Function '%s' lacks documentation", x.Name.Name),
					Suggestion: "Add a comment describing the function's purpose",
					Confidence: 0.8,
				})
			}
		}
		return true
	})

	return suggestions, nil
}

// AnalyzePackage analyzes all files in a package
func (my *CodeAnalyzer) AnalyzePackage(dir string) (map[string][]Suggestion, error) {
	files, err := my.accessor.ListGoFiles(dir)
	if err != nil {
		return nil, err
	}

	results := make(map[string][]Suggestion)
	for _, file := range files {
		suggestions, err := my.AnalyzeFile(file)
		if err != nil {
			continue
		}
		results[file] = suggestions
	}

	return results, nil
}
