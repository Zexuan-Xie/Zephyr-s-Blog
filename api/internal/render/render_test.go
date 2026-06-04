package render

import (
	"strings"
	"testing"
)

func TestMarkdownToSafeHTMLSanitizesUnsafeContent(t *testing.T) {
	safeHTML, searchText, err := MarkdownToSafeHTML(
		"# Hi & welcome\n\n" +
			"[unsafe](javascript:alert(1))\n\n" +
			"<img src=x onerror=alert(1)>\n" +
			"<script>alert(1)</script>",
	)
	if err != nil {
		t.Fatalf("MarkdownToSafeHTML() error = %v", err)
	}

	for _, unsafe := range []string{"<script", "javascript:", "onerror", "alert(1)"} {
		if strings.Contains(strings.ToLower(safeHTML), unsafe) {
			t.Fatalf("unsafe content %q survived in %q", unsafe, safeHTML)
		}
	}
	if searchText != "Hi & welcome unsafe" {
		t.Fatalf("searchText = %q, want %q", searchText, "Hi & welcome unsafe")
	}
}

func TestMarkdownToSafeHTMLProducesNormalizedPlainText(t *testing.T) {
	_, searchText, err := MarkdownToSafeHTML("One **bold**\n\n- two\n- three")
	if err != nil {
		t.Fatalf("MarkdownToSafeHTML() error = %v", err)
	}
	if searchText != "One bold two three" {
		t.Fatalf("searchText = %q, want %q", searchText, "One bold two three")
	}
}

func TestVisibleTextFromHTMLExcludesNonVisibleContent(t *testing.T) {
	document := `<html>
		<head>
			<title>Browser title</title>
			<style>.secret { display: block }</style>
			<script>secret()</script>
		</head>
		<body>
			<h1>Visible</h1>
			<p hidden>Hidden attribute</p>
			<p aria-hidden="TRUE">ARIA hidden</p>
			<p style="display: none">Inline hidden</p>
			<template>Template hidden</template>
			<noscript>Noscript hidden</noscript>
			<p>Text &amp; more</p>
		</body>
	</html>`

	if got := VisibleTextFromHTML(document); got != "Visible Text & more" {
		t.Fatalf("VisibleTextFromHTML() = %q, want %q", got, "Visible Text & more")
	}
}

func TestVisibleTextFromHTMLExcludesHiddenDescendantsAndImportantStyles(t *testing.T) {
	document := `<main>
		<section hidden><p>Nested hidden</p></section>
		<p style="DISPLAY: none !important"><span>Important hidden</span></p>
		<p style="visibility: hidden !IMPORTANT">Also hidden</p>
		<p aria-hidden="false">Visible</p>
	</main>`

	if got := VisibleTextFromHTML(document); got != "Visible" {
		t.Fatalf("VisibleTextFromHTML() = %q, want %q", got, "Visible")
	}
}

func TestVisibleTextFromHTMLNormalizesWhitespace(t *testing.T) {
	document := "<main><p>  first\n\tline </p><p>second&nbsp;line&#x2003;third</p></main>"

	if got := VisibleTextFromHTML(document); got != "first line second line third" {
		t.Fatalf("VisibleTextFromHTML() = %q, want %q", got, "first line second line third")
	}
}

func TestReadingTimeMinutesRoundsUpAndHasMinimum(t *testing.T) {
	if got := ReadingTimeMinutes(""); got != 1 {
		t.Fatalf("ReadingTimeMinutes(empty) = %d, want 1", got)
	}
	if got := ReadingTimeMinutes(strings.Repeat("word ", 220)); got != 1 {
		t.Fatalf("ReadingTimeMinutes(220 words) = %d, want 1", got)
	}
	if got := ReadingTimeMinutes(strings.Repeat("word ", 221)); got != 2 {
		t.Fatalf("ReadingTimeMinutes(221 words) = %d, want 2", got)
	}
}
