package filesystem

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

type Replacer func(content, find string) []string

func SimpleReplacer(content, find string) []string {
	if strings.Contains(content, find) {
		return []string{find}
	}
	return nil
}

func LineTrimmedReplacer(content, find string) []string {
	originalLines := strings.Split(content, "\n")
	searchLines := strings.Split(find, "\n")

	if len(searchLines) > 0 && searchLines[len(searchLines)-1] == "" {
		searchLines = searchLines[:len(searchLines)-1]
	}

	var results []string

	for i := 0; i <= len(originalLines)-len(searchLines); i++ {
		matches := true

		for j := 0; j < len(searchLines); j++ {
			originalTrimmed := strings.TrimSpace(originalLines[i+j])
			searchTrimmed := strings.TrimSpace(searchLines[j])

			if originalTrimmed != searchTrimmed {
				matches = false
				break
			}
		}

		if matches {
			matchStartIndex := 0
			for k := 0; k < i; k++ {
				matchStartIndex += len(originalLines[k]) + 1
			}

			matchEndIndex := matchStartIndex
			for k := 0; k < len(searchLines); k++ {
				matchEndIndex += len(originalLines[i+k])
				if k < len(searchLines)-1 {
					matchEndIndex += 1
				}
			}

			results = append(results, content[matchStartIndex:matchEndIndex])
		}
	}

	return results
}

func BlockAnchorReplacer(content, find string) []string {
	originalLines := strings.Split(content, "\n")
	searchLines := strings.Split(find, "\n")

	if len(searchLines) < 3 {
		return nil
	}

	if len(searchLines) > 0 && searchLines[len(searchLines)-1] == "" {
		searchLines = searchLines[:len(searchLines)-1]
	}

	firstLineSearch := strings.TrimSpace(searchLines[0])
	lastLineSearch := strings.TrimSpace(searchLines[len(searchLines)-1])
	searchBlockSize := len(searchLines)

	type candidate struct {
		startLine int
		endLine   int
	}

	var candidates []candidate
	for i := 0; i < len(originalLines); i++ {
		if strings.TrimSpace(originalLines[i]) != firstLineSearch {
			continue
		}

		for j := i + 2; j < len(originalLines); j++ {
			if strings.TrimSpace(originalLines[j]) == lastLineSearch {
				candidates = append(candidates, candidate{startLine: i, endLine: j})
				break
			}
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	const singleCandidateThreshold = 0.0
	const multipleCandidatesThreshold = 0.3

	if len(candidates) == 1 {
		c := candidates[0]
		actualBlockSize := c.endLine - c.startLine + 1

		similarity := 0.0
		linesToCheck := min(searchBlockSize-2, actualBlockSize-2)

		if linesToCheck > 0 {
			for j := 1; j < searchBlockSize-1 && j < actualBlockSize-1; j++ {
				originalLine := strings.TrimSpace(originalLines[c.startLine+j])
				searchLine := strings.TrimSpace(searchLines[j])
				maxLen := max(len(originalLine), len(searchLine))
				if maxLen == 0 {
					continue
				}
				distance := levenshtein(originalLine, searchLine)
				similarity += (1.0 - float64(distance)/float64(maxLen)) / float64(linesToCheck)

				if similarity >= singleCandidateThreshold {
					break
				}
			}
		} else {
			similarity = 1.0
		}

		if similarity >= singleCandidateThreshold {
			return []string{extractBlock(content, originalLines, c.startLine, c.endLine)}
		}
		return nil
	}

	var bestMatch *candidate
	maxSimilarity := -1.0

	for _, c := range candidates {
		actualBlockSize := c.endLine - c.startLine + 1

		similarity := 0.0
		linesToCheck := min(searchBlockSize-2, actualBlockSize-2)

		if linesToCheck > 0 {
			for j := 1; j < searchBlockSize-1 && j < actualBlockSize-1; j++ {
				originalLine := strings.TrimSpace(originalLines[c.startLine+j])
				searchLine := strings.TrimSpace(searchLines[j])
				maxLen := max(len(originalLine), len(searchLine))
				if maxLen == 0 {
					continue
				}
				distance := levenshtein(originalLine, searchLine)
				similarity += 1.0 - float64(distance)/float64(maxLen)
			}
			similarity /= float64(linesToCheck)
		} else {
			similarity = 1.0
		}

		if similarity > maxSimilarity {
			maxSimilarity = similarity
			c := c
			bestMatch = &c
		}
	}

	if maxSimilarity >= multipleCandidatesThreshold && bestMatch != nil {
		return []string{extractBlock(content, originalLines, bestMatch.startLine, bestMatch.endLine)}
	}

	return nil
}

func extractBlock(content string, lines []string, startLine, endLine int) string {
	matchStartIndex := 0
	for k := 0; k < startLine; k++ {
		matchStartIndex += len(lines[k]) + 1
	}
	matchEndIndex := matchStartIndex
	for k := startLine; k <= endLine; k++ {
		matchEndIndex += len(lines[k])
		if k < endLine {
			matchEndIndex += 1
		}
	}
	return content[matchStartIndex:matchEndIndex]
}

func WhitespaceNormalizedReplacer(content, find string) []string {
	normalizeWhitespace := func(text string) string {
		return strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(text, " "))
	}

	normalizedFind := normalizeWhitespace(find)
	var results []string

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if normalizeWhitespace(line) == normalizedFind {
			results = append(results, line)
		} else {
			normalizedLine := normalizeWhitespace(line)
			if strings.Contains(normalizedLine, normalizedFind) {
				words := strings.Fields(strings.TrimSpace(find))
				if len(words) > 0 {
					var patternParts []string
					for _, word := range words {
						patternParts = append(patternParts, regexp.QuoteMeta(word))
					}
					pattern := strings.Join(patternParts, `\s+`)
					re, err := regexp.Compile(pattern)
					if err == nil {
						if match := re.FindString(line); match != "" {
							results = append(results, match)
						}
					}
				}
			}
		}
	}

	findLines := strings.Split(find, "\n")
	if len(findLines) > 1 {
		for i := 0; i <= len(lines)-len(findLines); i++ {
			block := strings.Join(lines[i:i+len(findLines)], "\n")
			if normalizeWhitespace(block) == normalizedFind {
				results = append(results, block)
			}
		}
	}

	return results
}

func IndentationFlexibleReplacer(content, find string) []string {
	removeIndentation := func(text string) string {
		lines := strings.Split(text, "\n")
		var nonEmptyLines []string
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				nonEmptyLines = append(nonEmptyLines, line)
			}
		}
		if len(nonEmptyLines) == 0 {
			return text
		}

		minIndent := -1
		for _, line := range nonEmptyLines {
			indent := 0
			for _, c := range line {
				if c == ' ' || c == '\t' {
					indent++
				} else {
					break
				}
			}
			if minIndent == -1 || indent < minIndent {
				minIndent = indent
			}
		}

		var result []string
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				result = append(result, line)
			} else if len(line) >= minIndent {
				result = append(result, line[minIndent:])
			} else {
				result = append(result, line)
			}
		}
		return strings.Join(result, "\n")
	}

	normalizedFind := removeIndentation(find)
	contentLines := strings.Split(content, "\n")
	findLines := strings.Split(find, "\n")

	var results []string

	for i := 0; i <= len(contentLines)-len(findLines); i++ {
		block := strings.Join(contentLines[i:i+len(findLines)], "\n")
		if removeIndentation(block) == normalizedFind {
			results = append(results, block)
		}
	}

	return results
}

func EscapeNormalizedReplacer(content, find string) []string {
	unescapeString := func(str string) string {
		replacements := map[string]string{
			`\n`:  "\n",
			`\t`:  "\t",
			`\r`:  "\r",
			`\'`:  "'",
			`\"`:  "\"",
			"\\`": "`",
			`\\`:  "\\",
			`\$`:  "$",
		}
		result := str
		for escaped, unescaped := range replacements {
			result = strings.ReplaceAll(result, escaped, unescaped)
		}
		return result
	}

	unescapedFind := unescapeString(find)
	var results []string

	if strings.Contains(content, unescapedFind) {
		results = append(results, unescapedFind)
	}

	lines := strings.Split(content, "\n")
	findLines := strings.Split(unescapedFind, "\n")

	for i := 0; i <= len(lines)-len(findLines); i++ {
		block := strings.Join(lines[i:i+len(findLines)], "\n")
		unescapedBlock := unescapeString(block)

		if unescapedBlock == unescapedFind {
			results = append(results, block)
		}
	}

	return results
}

func MultiOccurrenceReplacer(content, find string) []string {
	var results []string
	startIndex := 0

	for {
		index := strings.Index(content[startIndex:], find)
		if index == -1 {
			break
		}
		results = append(results, find)
		startIndex += index + len(find)
	}

	return results
}

func TrimmedBoundaryReplacer(content, find string) []string {
	trimmedFind := strings.TrimSpace(find)

	if trimmedFind == find {
		return nil
	}

	var results []string

	if strings.Contains(content, trimmedFind) {
		results = append(results, trimmedFind)
	}

	lines := strings.Split(content, "\n")
	findLines := strings.Split(find, "\n")

	for i := 0; i <= len(lines)-len(findLines); i++ {
		block := strings.Join(lines[i:i+len(findLines)], "\n")

		if strings.TrimSpace(block) == trimmedFind {
			results = append(results, block)
		}
	}

	return results
}

func ContextAwareReplacer(content, find string) []string {
	findLines := strings.Split(find, "\n")
	if len(findLines) < 3 {
		return nil
	}

	if len(findLines) > 0 && findLines[len(findLines)-1] == "" {
		findLines = findLines[:len(findLines)-1]
	}

	contentLines := strings.Split(content, "\n")

	firstLine := strings.TrimSpace(findLines[0])
	lastLine := strings.TrimSpace(findLines[len(findLines)-1])

	var results []string

	for i := 0; i < len(contentLines); i++ {
		if strings.TrimSpace(contentLines[i]) != firstLine {
			continue
		}

		for j := i + 2; j < len(contentLines); j++ {
			if strings.TrimSpace(contentLines[j]) == lastLine {
				blockLines := contentLines[i : j+1]
				block := strings.Join(blockLines, "\n")

				if len(blockLines) == len(findLines) {
					matchingLines := 0
					totalNonEmptyLines := 0

					for k := 1; k < len(blockLines)-1; k++ {
						blockLine := strings.TrimSpace(blockLines[k])
						findLine := strings.TrimSpace(findLines[k])

						if len(blockLine) > 0 || len(findLine) > 0 {
							totalNonEmptyLines++
							if blockLine == findLine {
								matchingLines++
							}
						}
					}

					if totalNonEmptyLines == 0 || float64(matchingLines)/float64(totalNonEmptyLines) >= 0.5 {
						results = append(results, block)
						break
					}
				}
				break
			}
		}
	}

	return results
}

func levenshtein(a, b string) int {
	if a == "" || b == "" {
		return max(utf8.RuneCountInString(a), utf8.RuneCountInString(b))
	}

	aRunes := []rune(a)
	bRunes := []rune(b)
	aLen := len(aRunes)
	bLen := len(bRunes)

	matrix := make([][]int, aLen+1)
	for i := range matrix {
		matrix[i] = make([]int, bLen+1)
		matrix[i][0] = i
	}
	for j := 0; j <= bLen; j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= aLen; i++ {
		for j := 1; j <= bLen; j++ {
			cost := 0
			if aRunes[i-1] != bRunes[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}

	return matrix[aLen][bLen]
}

var AllReplacers = []Replacer{
	SimpleReplacer,
	LineTrimmedReplacer,
	BlockAnchorReplacer,
	WhitespaceNormalizedReplacer,
	IndentationFlexibleReplacer,
	EscapeNormalizedReplacer,
	TrimmedBoundaryReplacer,
	ContextAwareReplacer,
	MultiOccurrenceReplacer,
}

func FindReplacement(content, oldString string, replaceAll bool) (string, error) {
	for _, replacer := range AllReplacers {
		matches := replacer(content, oldString)
		if len(matches) == 0 {
			continue
		}

		search := matches[0]
		index := strings.Index(content, search)
		if index == -1 {
			continue
		}

		if replaceAll {
			return search, nil
		}

		lastIndex := strings.LastIndex(content, search)
		if index != lastIndex {
			continue
		}

		return search, nil
	}

	return "", nil
}
