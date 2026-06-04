package render

import (
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var whitespacePattern = regexp.MustCompile(`\s+`)
var hiddenStylePattern = regexp.MustCompile(`(?i)(?:^|;)\s*(?:display\s*:\s*none|visibility\s*:\s*hidden)\s*(?:;|$)`)

// VisibleTextFromHTML extracts normalized text that a document can visibly
// present. Executable, metadata, template, and hidden subtrees are excluded.
func VisibleTextFromHTML(document string) string {
	root, err := html.Parse(strings.NewReader(document))
	if err != nil {
		return ""
	}

	parts := make([]string, 0)
	walkVisibleText(root, false, &parts)
	return strings.Join(parts, " ")
}

func walkVisibleText(node *html.Node, hidden bool, parts *[]string) {
	if node.Type == html.ElementNode {
		if isNonVisibleElement(node.Data) {
			return
		}
		hidden = hidden || hasHiddenAttribute(node)
	}

	if node.Type == html.TextNode && !hidden {
		if text := normalizeWhitespace(node.Data); text != "" {
			*parts = append(*parts, text)
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walkVisibleText(child, hidden, parts)
	}
}

func isNonVisibleElement(name string) bool {
	switch strings.ToLower(name) {
	case "head", "script", "style", "template", "noscript", "meta", "link":
		return true
	default:
		return false
	}
}

func hasHiddenAttribute(node *html.Node) bool {
	for _, attribute := range node.Attr {
		key := strings.ToLower(attribute.Key)
		value := strings.TrimSpace(attribute.Val)
		switch key {
		case "hidden":
			return true
		case "aria-hidden":
			if strings.EqualFold(value, "true") {
				return true
			}
		case "style":
			if hiddenStylePattern.MatchString(value) {
				return true
			}
		}
	}
	return false
}

func normalizeWhitespace(text string) string {
	return strings.TrimSpace(whitespacePattern.ReplaceAllString(text, " "))
}
