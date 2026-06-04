package render

import "strings"

const wordsPerMinute = 220

// ReadingTimeMinutes estimates reading time, rounded up to at least one minute.
func ReadingTimeMinutes(text string) int {
	wordCount := len(strings.Fields(text))
	if wordCount == 0 {
		return 1
	}
	return (wordCount + wordsPerMinute - 1) / wordsPerMinute
}
