package render

import (
	"bytes"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

var markdownSanitizer = func() *bluemonday.Policy {
	policy := bluemonday.UGCPolicy()
	policy.RequireNoFollowOnLinks(false)
	return policy
}()

// MarkdownToSafeHTML renders Markdown into sanitized HTML and searchable plain
// text. Raw HTML is not enabled in goldmark, and the rendered result is
// sanitized before either value is returned.
func MarkdownToSafeHTML(markdown string) (safeHTML string, searchText string, err error) {
	var rendered bytes.Buffer
	if err := goldmark.Convert([]byte(markdown), &rendered); err != nil {
		return "", "", err
	}

	safeHTML = markdownSanitizer.Sanitize(rendered.String())
	return safeHTML, VisibleTextFromHTML(safeHTML), nil
}
