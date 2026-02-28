package service

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
)

// CodeChunker 代码分块器
type CodeChunker struct {
	maxChunkSize int // 最大块大小（字符数）
	overlap      int // 重叠大小（行数）
}

type Chunk struct {
	Content   string
	StartLine int
	EndLine   int
	Symbols   []string
}

type FileChunk struct {
	FilePath string
	Language string
	Chunks   []*Chunk
}

// NewCodeChunker 创建分块器
func NewCodeChunker(maxChunkSize, overlap int) *CodeChunker {
	if maxChunkSize <= 0 {
		maxChunkSize = 2000
	}
	if overlap <= 0 {
		overlap = 200
	}
	return &CodeChunker{maxChunkSize: maxChunkSize, overlap: overlap}
}

// ChunkFile 对文件进行分块
func (c *CodeChunker) ChunkFile(filePath, content string) *FileChunk {
	lang := c.detectLanguage(filePath)
	if lang == "go" {
		return c.chunkGoFile(filePath, content)
	}
	return c.chunkGenericFile(filePath, content, lang)
}

// detectLanguage 检测编程语言
func (c *CodeChunker) detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go":
		return "go"
	case ".js", ".jsx", ".ts", ".tsx":
		return "javascript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".rs":
		return "rust"
	case ".md":
		return "markdown"
	default:
		return "text"
	}
}

// chunkGoFile Go 文件按语法树分块
func (c *CodeChunker) chunkGoFile(filePath, content string) *FileChunk {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return c.chunkGenericFile(filePath, content, "go")
	}

	lines := strings.Split(content, "\n")
	chunks := make([]*Chunk, 0, len(f.Decls))

	for _, decl := range f.Decls {
		start := fset.Position(decl.Pos()).Line
		end := fset.Position(decl.End()).Line
		symbols := c.extractDeclSymbols(decl)

		var buf strings.Builder
		for i := start - 1; i < end && i < len(lines); i++ {
			buf.WriteString(lines[i])
			buf.WriteString("\n")
		}

		chunks = append(chunks, &Chunk{
			Content:   buf.String(),
			StartLine: start,
			EndLine:   end,
			Symbols:   symbols,
		})
	}

	if len(chunks) == 0 {
		return c.chunkGenericFile(filePath, content, "go")
	}

	return &FileChunk{FilePath: filePath, Language: "go", Chunks: chunks}
}

// extractDeclSymbols 提取声明中的符号
func (c *CodeChunker) extractDeclSymbols(decl ast.Decl) []string {
	symbols := make([]string, 0)

	switch d := decl.(type) {
	case *ast.FuncDecl:
		if d.Recv != nil {
			symbols = append(symbols, "method:"+d.Name.Name)
		} else {
			symbols = append(symbols, "func:"+d.Name.Name)
		}
	case *ast.GenDecl:
		switch d.Tok {
		case token.TYPE:
			for _, spec := range d.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					symbols = append(symbols, "type:"+ts.Name.Name)
				}
			}
		case token.CONST, token.VAR:
			for _, spec := range d.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range vs.Names {
						symbols = append(symbols, "var:"+name.Name)
					}
				}
			}
		}
	}

	return symbols
}

// chunkGenericFile 通用文件按大小分块（带重叠）
func (c *CodeChunker) chunkGenericFile(filePath, content, lang string) *FileChunk {
	lines := strings.Split(content, "\n")
	chunks := make([]*Chunk, 0)

	currentChunk := new(strings.Builder)
	startLine := 1
	overlapLines := make([]string, 0)

	for i, line := range lines {
		currentChunk.WriteString(line)
		currentChunk.WriteString("\n")

		if currentChunk.Len() >= c.maxChunkSize {
			chunks = append(chunks, &Chunk{
				Content:   currentChunk.String(),
				StartLine: startLine,
				EndLine:   i + 1,
				Symbols:   []string{},
			})

			overlapLines = c.getOverlapLines(strings.Split(currentChunk.String(), "\n"))

			currentChunk = new(strings.Builder)
			for _, ol := range overlapLines {
				currentChunk.WriteString(ol)
				currentChunk.WriteString("\n")
			}
			startLine = i + 1 - len(overlapLines)
			if startLine < 1 {
				startLine = 1
			}
		}
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, &Chunk{
			Content:   currentChunk.String(),
			StartLine: startLine,
			EndLine:   len(lines),
			Symbols:   []string{},
		})
	}

	return &FileChunk{FilePath: filePath, Language: lang, Chunks: chunks}
}

// getOverlapLines 获取重叠行
func (c *CodeChunker) getOverlapLines(lines []string) []string {
	if len(lines) <= c.overlap {
		return lines
	}
	return lines[len(lines)-c.overlap:]
}

// ExtractSymbols 从代码中提取符号（简单正则匹配）
func (c *CodeChunker) ExtractSymbols(content, lang string) []string {
	symbols := make([]string, 0)

	funcPatterns := map[string]*regexp.Regexp{
		"go":         regexp.MustCompile(`func\s+(\w+)`),
		"javascript": regexp.MustCompile(`function\s+(\w+)|const\s+(\w+)\s*=\s*\(|(\w+)\s*:\s*\(.*\)\s*=>`),
		"python":     regexp.MustCompile(`def\s+(\w+)`),
		"rust":       regexp.MustCompile(`fn\s+(\w+)`),
		"java":       regexp.MustCompile(`(public|private|protected)?\s*(static)?\s*\w+\s+(\w+)\s*\(`),
	}

	if pattern, ok := funcPatterns[lang]; ok {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			for i := 1; i < len(m); i++ {
				if m[i] != "" {
					symbols = append(symbols, m[i])
				}
			}
		}
	}

	return symbols
}
