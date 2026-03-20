package text

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	gmtext "github.com/yuin/goldmark/text"
)

var mdParser = goldmark.New(
	goldmark.WithExtensions(extension.Strikethrough),
).Parser()

// ConvertMarkdownToSlackMrkdwn converts standard markdown text to Slack's
// mrkdwn format. It always returns valid output (never errors).
func ConvertMarkdownToSlackMrkdwn(markdown string) string {
	if strings.TrimSpace(markdown) == "" {
		return ""
	}

	source := []byte(markdown)
	doc := mdParser.Parse(gmtext.NewReader(source))

	var blocks []string
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		rendered := renderBlock(child, source)
		if rendered != "" {
			blocks = append(blocks, rendered)
		}
	}

	return strings.TrimSpace(strings.Join(blocks, "\n\n"))
}

func renderBlock(node ast.Node, source []byte) string {
	switch node.Kind() {
	case ast.KindParagraph:
		return renderInlineChildren(node, source)

	case ast.KindHeading:
		inline := renderInlineChildren(node, source)
		return "*" + inline + "*"

	case ast.KindFencedCodeBlock, ast.KindCodeBlock:
		return renderCodeBlock(node, source)

	case ast.KindBlockquote:
		return renderBlockquote(node, source)

	case ast.KindList:
		return renderList(node, source)

	case ast.KindThematicBreak:
		return "---"

	default:
		// Unknown block: try inline children as fallback
		t := renderInlineChildren(node, source)
		if t != "" {
			return t
		}
		return ""
	}
}

func renderCodeBlock(node ast.Node, source []byte) string {
	var sb strings.Builder
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		sb.Write(line.Value(source))
	}
	code := strings.TrimRight(sb.String(), "\n")
	return "```\n" + code + "\n```"
}

func renderBlockquote(node ast.Node, source []byte) string {
	var parts []string
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		rendered := renderBlock(child, source)
		if rendered != "" {
			lines := strings.Split(rendered, "\n")
			for i, line := range lines {
				lines[i] = "> " + line
			}
			parts = append(parts, strings.Join(lines, "\n"))
		}
	}
	return strings.Join(parts, "\n")
}

func renderList(node ast.Node, source []byte) string {
	list := node.(*ast.List)
	var items []string
	counter := list.Start
	if counter == 0 {
		counter = 1
	}

	for item := node.FirstChild(); item != nil; item = item.NextSibling() {
		if item.Kind() != ast.KindListItem {
			continue
		}

		var itemParts []string
		for child := item.FirstChild(); child != nil; child = child.NextSibling() {
			rendered := renderBlock(child, source)
			if rendered != "" {
				itemParts = append(itemParts, rendered)
			}
		}

		itemText := strings.Join(itemParts, "\n")

		var prefix string
		if list.IsOrdered() {
			prefix = fmt.Sprintf("%d. ", counter)
			counter++
		} else {
			prefix = "\u2022 " // bullet: •
		}

		items = append(items, prefix+itemText)
	}
	return strings.Join(items, "\n")
}

func renderInlineChildren(node ast.Node, source []byte) string {
	var sb strings.Builder
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		renderInline(child, source, &sb)
	}
	return strings.TrimSpace(sb.String())
}

func renderInline(node ast.Node, source []byte, sb *strings.Builder) {
	switch node.Kind() {
	case ast.KindText:
		t := node.(*ast.Text)
		sb.Write(t.Segment.Value(source))
		if t.SoftLineBreak() {
			sb.WriteByte('\n')
		}
		if t.HardLineBreak() {
			sb.WriteByte('\n')
		}

	case ast.KindEmphasis:
		emp := node.(*ast.Emphasis)
		switch emp.Level {
		case 2:
			sb.WriteByte('*')
			renderInlineChildrenTo(node, source, sb)
			sb.WriteByte('*')
		case 1:
			sb.WriteByte('_')
			renderInlineChildrenTo(node, source, sb)
			sb.WriteByte('_')
		}

	case ast.KindCodeSpan:
		sb.WriteByte('`')
		for c := node.FirstChild(); c != nil; c = c.NextSibling() {
			if c.Kind() == ast.KindText {
				t := c.(*ast.Text)
				sb.Write(t.Segment.Value(source))
			}
		}
		sb.WriteByte('`')

	case ast.KindLink:
		link := node.(*ast.Link)
		var linkText strings.Builder
		renderInlineChildrenTo(node, source, &linkText)
		text := linkText.String()
		url := string(link.Destination)
		if text != "" && text != url {
			fmt.Fprintf(sb, "<%s|%s>", url, text)
		} else {
			fmt.Fprintf(sb, "<%s>", url)
		}

	case ast.KindAutoLink:
		autoLink := node.(*ast.AutoLink)
		sb.Write(autoLink.URL(source))

	case ast.KindImage:
		img := node.(*ast.Image)
		var altText strings.Builder
		renderInlineChildrenTo(node, source, &altText)
		text := altText.String()
		if text != "" {
			fmt.Fprintf(sb, "<%s|%s>", string(img.Destination), text)
		} else {
			fmt.Fprintf(sb, "<%s>", string(img.Destination))
		}

	case ast.KindRawHTML:
		// Strip HTML — Slack cannot render it

	default:
		if node.Kind() == extast.KindStrikethrough {
			sb.WriteByte('~')
			renderInlineChildrenTo(node, source, sb)
			sb.WriteByte('~')
		} else {
			// Unknown inline: recurse
			renderInlineChildrenTo(node, source, sb)
		}
	}
}

func renderInlineChildrenTo(node ast.Node, source []byte, sb *strings.Builder) {
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		renderInline(c, source, sb)
	}
}
