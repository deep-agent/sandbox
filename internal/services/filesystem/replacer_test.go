package filesystem

import (
	"testing"
)

func TestSimpleReplacer(t *testing.T) {
	tests := []struct {
		name    string
		content string
		find    string
		want    int
	}{
		{
			name:    "exact match",
			content: "hello world",
			find:    "hello",
			want:    1,
		},
		{
			name:    "no match",
			content: "hello world",
			find:    "foo",
			want:    0,
		},
		{
			name:    "full content match",
			content: "hello",
			find:    "hello",
			want:    1,
		},
		{
			name:    "multiline match",
			content: "line1\nline2\nline3",
			find:    "line2",
			want:    1,
		},
		{
			name:    "partial match",
			content: "hello world",
			find:    "lo wo",
			want:    1,
		},
		{
			name:    "empty find",
			content: "hello",
			find:    "",
			want:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SimpleReplacer(tt.content, tt.find)
			if len(result) != tt.want {
				t.Errorf("SimpleReplacer() returned %d matches, want %d", len(result), tt.want)
			}
			if tt.want > 0 && result[0] != tt.find {
				t.Errorf("SimpleReplacer() match = %q, want %q", result[0], tt.find)
			}
		})
	}
}

func TestLineTrimmedReplacer(t *testing.T) {
	tests := []struct {
		name    string
		content string
		find    string
		want    int
	}{
		{
			name:    "exact match single line",
			content: "hello world",
			find:    "hello world",
			want:    1,
		},
		{
			name:    "trimmed match single line",
			content: "  hello world  ",
			find:    "hello world",
			want:    1,
		},
		{
			name:    "multiline exact match",
			content: "line1\nline2\nline3",
			find:    "line1\nline2",
			want:    1,
		},
		{
			name:    "multiline trimmed match",
			content: "  line1  \n  line2  \nline3",
			find:    "line1\nline2",
			want:    1,
		},
		{
			name:    "no match",
			content: "hello world",
			find:    "foo bar",
			want:    0,
		},
		{
			name:    "match with trailing newline in find",
			content: "line1\nline2\nline3",
			find:    "line1\nline2\n",
			want:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LineTrimmedReplacer(tt.content, tt.find)
			if len(result) != tt.want {
				t.Errorf("LineTrimmedReplacer() returned %d matches, want %d", len(result), tt.want)
			}
		})
	}
}

func TestLineTrimmedReplacerPreservesOriginal(t *testing.T) {
	content := "  hello world  \n  foo bar  "
	find := "hello world\nfoo bar"

	result := LineTrimmedReplacer(content, find)
	if len(result) != 1 {
		t.Fatalf("LineTrimmedReplacer() returned %d matches, want 1", len(result))
	}

	if result[0] != "  hello world  \n  foo bar  " {
		t.Errorf("LineTrimmedReplacer() should return original content with whitespace, got %q", result[0])
	}
}

func TestFindReplacementSimple(t *testing.T) {
	content := "hello world"
	oldString := "hello"

	result, err := FindReplacement(content, oldString, false)
	if err != nil {
		t.Fatalf("FindReplacement() error = %v", err)
	}

	if result != "hello" {
		t.Errorf("FindReplacement() = %q, want %q", result, "hello")
	}
}

func TestFindReplacementNoMatch(t *testing.T) {
	content := "hello world"
	oldString := "foo"

	result, err := FindReplacement(content, oldString, false)
	if err != nil {
		t.Fatalf("FindReplacement() error = %v", err)
	}

	if result != "" {
		t.Errorf("FindReplacement() = %q, want empty string", result)
	}
}

func TestFindReplacementMultipleOccurrences(t *testing.T) {
	content := "hello hello hello"
	oldString := "hello"

	result, err := FindReplacement(content, oldString, false)
	if err != nil {
		t.Fatalf("FindReplacement() error = %v", err)
	}

	if result != "" {
		t.Errorf("FindReplacement() with multiple occurrences should return empty, got %q", result)
	}
}

func TestFindReplacementMultipleOccurrencesReplaceAll(t *testing.T) {
	content := "hello hello hello"
	oldString := "hello"

	result, err := FindReplacement(content, oldString, true)
	if err != nil {
		t.Fatalf("FindReplacement() error = %v", err)
	}

	if result != "hello" {
		t.Errorf("FindReplacement() with replaceAll = %q, want %q", result, "hello")
	}
}

func TestFindReplacementTrimmed(t *testing.T) {
	content := "  hello world  \nfoo bar"
	oldString := "hello world\nfoo bar"

	result, err := FindReplacement(content, oldString, false)
	if err != nil {
		t.Fatalf("FindReplacement() error = %v", err)
	}

	if result == "" {
		t.Error("FindReplacement() should find trimmed match")
	}
}

func TestFindReplacementMultiline(t *testing.T) {
	content := "func main() {\n\tfmt.Println(\"hello\")\n}"
	oldString := "func main() {\n\tfmt.Println(\"hello\")\n}"

	result, err := FindReplacement(content, oldString, false)
	if err != nil {
		t.Fatalf("FindReplacement() error = %v", err)
	}

	if result == "" {
		t.Error("FindReplacement() should find exact multiline match")
	}
}

func TestFindReplacementIndentationDifference(t *testing.T) {
	content := "    line1\n    line2"
	oldString := "line1\nline2"

	result, err := FindReplacement(content, oldString, false)
	if err != nil {
		t.Fatalf("FindReplacement() error = %v", err)
	}

	if result == "" {
		t.Error("FindReplacement() should find match with different indentation")
	}
}

func TestSimpleReplacerMultipleOccurrences(t *testing.T) {
	content := "hello world hello"
	find := "hello"

	result := SimpleReplacer(content, find)

	if len(result) != 1 {
		t.Errorf("SimpleReplacer() returned %d matches, want 1", len(result))
	}
}

func TestLineTrimmedReplacerMultipleMatches(t *testing.T) {
	content := "hello\nworld\nhello\nworld"
	find := "hello\nworld"

	result := LineTrimmedReplacer(content, find)

	if len(result) != 2 {
		t.Errorf("LineTrimmedReplacer() returned %d matches, want 2", len(result))
	}
}

func TestFindReplacementUniqueMatch(t *testing.T) {
	content := "line1\nline2\nline3"
	oldString := "line2"

	result, err := FindReplacement(content, oldString, false)
	if err != nil {
		t.Fatalf("FindReplacement() error = %v", err)
	}

	if result != "line2" {
		t.Errorf("FindReplacement() = %q, want %q", result, "line2")
	}
}

func TestAllReplacersExist(t *testing.T) {
	if len(AllReplacers) == 0 {
		t.Error("AllReplacers should not be empty")
	}

	content := "hello world"
	find := "hello"

	foundMatch := false
	for _, replacer := range AllReplacers {
		result := replacer(content, find)
		if len(result) > 0 {
			foundMatch = true
			break
		}
	}

	if !foundMatch {
		t.Error("At least one replacer should find a match for 'hello' in 'hello world'")
	}
}
