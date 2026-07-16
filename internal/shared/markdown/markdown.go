package markdown

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	exast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// TOCItem 目录项
type TOCItem struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	ID    string `json:"id"`
}

// engine 仅用于解析，不使用默认 HTML renderer。
var engine = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
)

// ToHTML 将 Markdown 转为统一结构化的 HTML。
func ToHTML(src string) (string, error) {
	reader := text.NewReader([]byte(src))
	doc := engine.Parser().Parse(reader)

	var buf bytes.Buffer
	_, err := walk(&buf, []byte(src), doc)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ToHTMLWithTOC 将 Markdown 转为 HTML 并生成目录。
func ToHTMLWithTOC(src string) (string, []TOCItem, error) {
	reader := text.NewReader([]byte(src))
	doc := engine.Parser().Parse(reader)

	var buf bytes.Buffer
	var toc []TOCItem
	idCounter := make(map[string]int)

	_, err := walkWithTOC(&buf, []byte(src), doc, &toc, idCounter)
	if err != nil {
		return "", nil, err
	}
	return buf.String(), toc, nil
}

// GenerateTOC 仅生成目录，不转换 HTML。
func GenerateTOC(src string) ([]TOCItem, error) {
	source := []byte(src)
	reader := text.NewReader(source)
	doc := engine.Parser().Parse(reader)

	var toc []TOCItem
	idCounter := make(map[string]int)

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if heading, ok := n.(*ast.Heading); ok {
				if heading.Level >= 1 && heading.Level <= 6 {
					text := strings.TrimSpace(string(heading.Text(source)))
					if text != "" {
						id := generateUniqueHeadingID(text, idCounter)
						toc = append(toc, TOCItem{
							Level: heading.Level,
							Text:  text,
							ID:    id,
						})
					}
				}
			}
		}
		return ast.WalkContinue, nil
	})
	return toc, nil
}

// generateHeadingID 生成标题 ID（用于锚点跳转）
func generateHeadingID(text string) string {
	id := strings.ToLower(text)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")
	id = strings.ReplaceAll(id, "(", "")
	id = strings.ReplaceAll(id, ")", "")
	id = strings.ReplaceAll(id, "[", "")
	id = strings.ReplaceAll(id, "]", "")
	id = strings.ReplaceAll(id, "{", "")
	id = strings.ReplaceAll(id, "}", "")
	id = strings.ReplaceAll(id, ":", "")
	id = strings.ReplaceAll(id, ";", "")
	id = strings.ReplaceAll(id, ",", "")
	id = strings.ReplaceAll(id, ".", "")
	id = strings.ReplaceAll(id, "?", "")
	id = strings.ReplaceAll(id, "!", "")
	id = strings.ReplaceAll(id, "@", "")
	id = strings.ReplaceAll(id, "#", "")
	id = strings.ReplaceAll(id, "$", "")
	id = strings.ReplaceAll(id, "%", "")
	id = strings.ReplaceAll(id, "^", "")
	id = strings.ReplaceAll(id, "&", "")
	id = strings.ReplaceAll(id, "*", "")
	id = strings.ReplaceAll(id, "+", "")
	id = strings.ReplaceAll(id, "=", "")
	id = strings.ReplaceAll(id, "|", "")
	id = strings.ReplaceAll(id, "\\", "")
	id = strings.ReplaceAll(id, "/", "")
	id = strings.ReplaceAll(id, "<", "")
	id = strings.ReplaceAll(id, ">", "")
	id = strings.ReplaceAll(id, "\"", "")
	id = strings.ReplaceAll(id, "'", "")
	id = strings.ReplaceAll(id, "`", "")
	id = strings.ReplaceAll(id, "~", "")
	id = strings.ReplaceAll(id, "——", "-")
	id = strings.ReplaceAll(id, "—", "-")
	id = strings.ReplaceAll(id, "–", "-")
	for strings.Contains(id, "--") {
		id = strings.ReplaceAll(id, "--", "-")
	}
	if strings.HasPrefix(id, "-") {
		id = id[1:]
	}
	if strings.HasSuffix(id, "-") {
		id = id[:len(id)-1]
	}
	return id
}

// generateUniqueHeadingID 生成唯一的标题 ID，处理重复标题
func generateUniqueHeadingID(text string, counter map[string]int) string {
	baseID := generateHeadingID(text)
	if count, exists := counter[baseID]; exists {
		counter[baseID] = count + 1
		return fmt.Sprintf("%s-%d", baseID, count+1)
	}
	counter[baseID] = 1
	return baseID
}

// walkWithTOC 递归遍历 AST 并渲染 HTML，同时收集目录项。
func walkWithTOC(w *bytes.Buffer, source []byte, node ast.Node, toc *[]TOCItem, idCounter map[string]int) (ast.WalkStatus, error) {
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		if err := renderNodeWithTOC(w, source, c, toc, idCounter); err != nil {
			return ast.WalkStop, err
		}
	}
	return ast.WalkContinue, nil
}

// renderNodeWithTOC 根据节点类型渲染开始/内容/结束标签，同时收集目录项。
func renderNodeWithTOC(w *bytes.Buffer, source []byte, node ast.Node, toc *[]TOCItem, idCounter map[string]int) error {
	switch n := node.(type) {
	case *ast.Document:
		return renderChildrenWithTOC(w, source, n, toc, idCounter)
	case *ast.Heading:
		return renderHeadingWithTOC(w, source, n, toc, idCounter)
	case *ast.Paragraph:
		return renderParagraph(w, source, n)
	case *ast.FencedCodeBlock, *ast.CodeBlock:
		return renderCodeBlock(w, source, n)
	case *ast.Blockquote:
		return renderBlockquote(w, source, n)
	case *ast.List:
		return renderList(w, source, n)
	case *ast.ListItem:
		return renderListItem(w, source, n)
	case *ast.ThematicBreak:
		_, _ = w.WriteString(`<hr class="md-hr" />`)
		return nil
	case *ast.HTMLBlock:
		return renderHTMLBlock(w, source, n)
	case *ast.Text:
		return renderText(w, source, n)
	case *ast.String:
		_, _ = w.Write(n.Value)
		return nil
	case *ast.Emphasis:
		return renderEmphasis(w, source, n)
	case *ast.Link:
		return renderLink(w, source, n)
	case *ast.Image:
		return renderImage(w, source, n)
	case *ast.CodeSpan:
		return renderCodeSpan(w, source, n)
	case *ast.AutoLink:
		return renderAutoLink(w, source, n)
	case *exast.Table:
		return renderTable(w, source, n)
	case *exast.TableHeader:
		return renderTableSection(w, source, n, "thead")
	case *exast.TableRow:
		return renderTableRow(w, source, n)
	case *exast.TableCell:
		return renderTableCell(w, source, n, false)
	case *ast.RawHTML:
		return renderRawHTML(w, source, n)
	default:
		return renderChildrenWithTOC(w, source, n, toc, idCounter)
	}
}

// renderChildrenWithTOC 遍历子节点并渲染，同时收集目录项。
func renderChildrenWithTOC(w *bytes.Buffer, source []byte, node ast.Node, toc *[]TOCItem, idCounter map[string]int) error {
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		if err := renderNodeWithTOC(w, source, c, toc, idCounter); err != nil {
			return err
		}
	}
	return nil
}

// renderHeadingWithTOC 渲染标题并收集目录项。
func renderHeadingWithTOC(w *bytes.Buffer, source []byte, n *ast.Heading, toc *[]TOCItem, idCounter map[string]int) error {
	tag := "h" + strconv.Itoa(n.Level)
	text := strings.TrimSpace(string(n.Text(source)))
	id := generateUniqueHeadingID(text, idCounter)

	*toc = append(*toc, TOCItem{
		Level: n.Level,
		Text:  text,
		ID:    id,
	})

	_, _ = w.WriteString("<")
	_, _ = w.WriteString(tag)
	_, _ = w.WriteString(` class="md-h md-`)
	_, _ = w.WriteString(tag)
	_, _ = w.WriteString(`" id="`)
	_, _ = w.WriteString(id)
	_, _ = w.WriteString(`">`)
	if err := renderChildren(w, source, n); err != nil {
		return err
	}
	close(w, tag)
	return nil
}

// walk 递归遍历 AST 并渲染 HTML。
func walk(w *bytes.Buffer, source []byte, node ast.Node) (ast.WalkStatus, error) {
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		if err := renderNode(w, source, c); err != nil {
			return ast.WalkStop, err
		}
	}
	return ast.WalkContinue, nil
}

// renderNode 根据节点类型渲染开始/内容/结束标签。
func renderNode(w *bytes.Buffer, source []byte, node ast.Node) error {
	switch n := node.(type) {
	case *ast.Document:
		return renderChildren(w, source, n)
	case *ast.Heading:
		return renderHeading(w, source, n)
	case *ast.Paragraph:
		return renderParagraph(w, source, n)
	case *ast.FencedCodeBlock, *ast.CodeBlock:
		return renderCodeBlock(w, source, n)
	case *ast.Blockquote:
		return renderBlockquote(w, source, n)
	case *ast.List:
		return renderList(w, source, n)
	case *ast.ListItem:
		return renderListItem(w, source, n)
	case *ast.ThematicBreak:
		_, _ = w.WriteString(`<hr class="md-hr" />`)
		return nil
	case *ast.HTMLBlock:
		return renderHTMLBlock(w, source, n)
	case *ast.Text:
		return renderText(w, source, n)
	case *ast.String:
		_, _ = w.Write(n.Value)
		return nil
	case *ast.Emphasis:
		return renderEmphasis(w, source, n)
	case *ast.Link:
		return renderLink(w, source, n)
	case *ast.Image:
		return renderImage(w, source, n)
	case *ast.CodeSpan:
		return renderCodeSpan(w, source, n)
	case *ast.AutoLink:
		return renderAutoLink(w, source, n)
	case *exast.Table:
		return renderTable(w, source, n)
	case *exast.TableHeader:
		return renderTableSection(w, source, n, "thead")
	case *exast.TableRow:
		return renderTableRow(w, source, n)
	case *exast.TableCell:
		return renderTableCell(w, source, n, false)
	case *ast.RawHTML:
		return renderRawHTML(w, source, n)
	default:
		return renderChildren(w, source, n)
	}
}

// renderChildren 遍历子节点并渲染。
func renderChildren(w *bytes.Buffer, source []byte, node ast.Node) error {
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		if err := renderNode(w, source, c); err != nil {
			return err
		}
	}
	return nil
}

// open 写入带 class 的开始标签。
func open(w *bytes.Buffer, name, class string) {
	_, _ = w.WriteString("<")
	_, _ = w.WriteString(name)
	if class != "" {
		_, _ = w.WriteString(` class="`)
		_, _ = w.WriteString(class)
		_, _ = w.WriteString(`"`)
	}
	_, _ = w.WriteString(">")
}

// close 写入关闭标签。
func close(w *bytes.Buffer, name string) {
	_, _ = w.WriteString("</")
	_, _ = w.WriteString(name)
	_, _ = w.WriteString(">")
}

func renderHeading(w *bytes.Buffer, source []byte, n *ast.Heading) error {
	tag := "h" + strconv.Itoa(n.Level)
	open(w, tag, "md-h md-"+tag)
	if err := renderChildren(w, source, n); err != nil {
		return err
	}
	close(w, tag)
	return nil
}

func renderParagraph(w *bytes.Buffer, source []byte, n *ast.Paragraph) error {
	open(w, "p", "md-p")
	if err := renderChildren(w, source, n); err != nil {
		return err
	}
	close(w, "p")
	return nil
}

func renderBlockquote(w *bytes.Buffer, source []byte, n *ast.Blockquote) error {
	open(w, "blockquote", "md-blockquote")
	if err := renderChildren(w, source, n); err != nil {
		return err
	}
	close(w, "blockquote")
	return nil
}

func renderList(w *bytes.Buffer, source []byte, n *ast.List) error {
	if n.IsOrdered() {
		open(w, "ol", "md-ol")
	} else {
		open(w, "ul", "md-ul")
	}
	if err := renderChildren(w, source, n); err != nil {
		return err
	}
	if n.IsOrdered() {
		close(w, "ol")
	} else {
		close(w, "ul")
	}
	return nil
}

func renderListItem(w *bytes.Buffer, source []byte, n *ast.ListItem) error {
	open(w, "li", "md-li")
	if err := renderChildren(w, source, n); err != nil {
		return err
	}
	close(w, "li")
	return nil
}

func renderCodeBlock(w *bytes.Buffer, source []byte, node ast.Node) error {
	var lang string
	if cb, ok := node.(*ast.FencedCodeBlock); ok {
		if cb.Language(source) != nil {
			lang = string(cb.Language(source))
		}
	}

	var lines []string
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		raw := strings.TrimRight(string(line.Value(source)), "\r\n")
		lines = append(lines, raw)
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	_, _ = w.WriteString(`<pre class="md-code-block"`)
	if lang != "" {
		_, _ = w.WriteString(` data-lang="`)
		_, _ = w.WriteString(lang)
		_, _ = w.WriteString(`"`)
	}
	_, _ = w.WriteString(">")

	if lang != "" {
		_, _ = w.WriteString(`<div class="md-code-header"><span class="md-code-lang">`)
		_, _ = w.WriteString(lang)
		_, _ = w.WriteString(`</span></div>`)
	}

	_, _ = w.WriteString(`<div class="md-code-body">`)
	for i, line := range lines {
		_, _ = w.WriteString(`<div class="md-code-row"><span class="md-line-num">`)
		_, _ = w.WriteString(strconv.Itoa(i + 1))
		_, _ = w.WriteString(`</span><code class="md-code-line">`)
		_, _ = w.Write(util.EscapeHTML([]byte(line)))
		_, _ = w.WriteString(`</code></div>`)
	}
	_, _ = w.WriteString(`</div></pre>`)

	return nil
}

func renderHTMLBlock(w *bytes.Buffer, source []byte, n *ast.HTMLBlock) error {
	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		_, _ = w.Write(line.Value(source))
	}
	return nil
}

func renderText(w *bytes.Buffer, source []byte, n *ast.Text) error {
	seg := n.Segment
	val := seg.Value(source)
	if n.HardLineBreak() || n.SoftLineBreak() {
		val = bytes.TrimRight(val, "\r\n")
	}
	_, _ = w.Write(util.EscapeHTML(val))
	if n.HardLineBreak() {
		_, _ = w.WriteString(`<br class="md-br" />`)
	} else if n.SoftLineBreak() {
		_, _ = w.WriteString(" ")
	}
	return nil
}

func renderEmphasis(w *bytes.Buffer, source []byte, n *ast.Emphasis) error {
	if n.Level == 2 {
		open(w, "strong", "md-strong")
	} else {
		open(w, "em", "md-em")
	}
	if err := renderChildren(w, source, n); err != nil {
		return err
	}
	if n.Level == 2 {
		close(w, "strong")
	} else {
		close(w, "em")
	}
	return nil
}

func renderLink(w *bytes.Buffer, source []byte, n *ast.Link) error {
	_, _ = w.WriteString(`<a class="md-link" href="`)
	_, _ = w.Write(n.Destination)
	_, _ = w.WriteString(`"`)
	if n.Title != nil && len(n.Title) > 0 {
		_, _ = w.WriteString(` title="`)
		_, _ = w.Write(n.Title)
		_, _ = w.WriteString(`"`)
	}
	_, _ = w.WriteString(">")
	if err := renderChildren(w, source, n); err != nil {
		return err
	}
	close(w, "a")
	return nil
}

func renderImage(w *bytes.Buffer, source []byte, n *ast.Image) error {
	_, _ = w.WriteString(`<img class="md-img" src="`)
	_, _ = w.Write(n.Destination)
	_, _ = w.WriteString(`" alt="`)
	_, _ = w.Write(n.Text(source))
	_, _ = w.WriteString(`" />`)
	return nil
}

func renderCodeSpan(w *bytes.Buffer, source []byte, n *ast.CodeSpan) error {
	open(w, "code", "md-inline-code")
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if s, ok := c.(*ast.String); ok {
			_, _ = w.Write(util.EscapeHTML(s.Value))
		}
	}
	close(w, "code")
	return nil
}

// ---------- 表格渲染 ----------

func renderTable(w *bytes.Buffer, source []byte, n *exast.Table) error {
	open(w, "table", "md-table")
	if err := renderChildren(w, source, n); err != nil {
		return err
	}
	close(w, "table")
	return nil
}

func renderTableSection(w *bytes.Buffer, source []byte, n ast.Node, tag string) error {
	open(w, tag, "")
	if err := renderChildren(w, source, n); err != nil {
		return err
	}
	close(w, tag)
	return nil
}

func renderTableRow(w *bytes.Buffer, source []byte, n *exast.TableRow) error {
	open(w, "tr", "md-tr")
	if err := renderChildren(w, source, n); err != nil {
		return err
	}
	close(w, "tr")
	return nil
}

func renderTableCell(w *bytes.Buffer, source []byte, n *exast.TableCell, _ bool) error {
	tag := "td"
	class := "md-td"
	// Cell → Row → TableHeader，检查祖父节点是否为 TableHeader
	if row := n.Parent(); row != nil {
		if section := row.Parent(); section != nil {
			if _, ok := section.(*exast.TableHeader); ok {
				tag = "th"
				class = "md-th"
			}
		}
	}
	open(w, tag, class)
	if err := renderChildren(w, source, n); err != nil {
		return err
	}
	close(w, tag)
	return nil
}

func renderAutoLink(w *bytes.Buffer, source []byte, n *ast.AutoLink) error {
	_, _ = w.WriteString(`<a class="md-link md-autolink" href="`)
	_, _ = w.Write(n.URL(source))
	_, _ = w.WriteString(`">`)
	_, _ = w.Write(n.Label(source))
	_, _ = w.WriteString("</a>")
	return nil
}

func renderRawHTML(w *bytes.Buffer, source []byte, n *ast.RawHTML) error {
	for i := 0; i < n.Segments.Len(); i++ {
		seg := n.Segments.At(i)
		_, _ = w.Write(seg.Value(source))
	}
	return nil
}
