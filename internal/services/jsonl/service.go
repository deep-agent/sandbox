package jsonl

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const maxScanBufferSize = 10 * 1024 * 1024 // 10MB

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CountLines(file string) (int, error) {
	f, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxScanBufferSize), maxScanBufferSize)
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) ReadLines(file string, startLine int, count *int) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxScanBufferSize), maxScanBufferSize)

	// Skip lines before startLine
	currentLine := 0
	for currentLine < startLine && scanner.Scan() {
		currentLine++
	}

	// Read requested lines
	lines := make([]string, 0)
	if count == nil {
		// Read until end of file
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
	} else {
		// Read up to count lines
		for len(lines) < *count && scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func (s *Service) AppendLine(file string, jsonStrings []string) error {
	for i, s := range jsonStrings {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("json_string[%d] is empty", i)
		}
	}

	// Check if existing file needs a leading newline
	info, err := os.Stat(file)
	if err == nil && info.Size() > 0 {
		rf, err := os.Open(file)
		if err != nil {
			return err
		}
		buf := make([]byte, 1)
		_, err = rf.ReadAt(buf, info.Size()-1)
		rf.Close()
		if err != nil {
			return err
		}
		if buf[0] != '\n' {
			f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			_, err = f.WriteString("\n")
			f.Close()
			if err != nil {
				return err
			}
		}
	}

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, s := range jsonStrings {
		if strings.HasSuffix(s, "\n") {
			if _, err := w.WriteString(s); err != nil {
				return err
			}
			continue
		}
		if _, err := w.WriteString(s + "\n"); err != nil {
			return err
		}
	}

	return w.Flush()
}
