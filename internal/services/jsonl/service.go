package jsonl

import (
	"bufio"
	"os"
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

func (s *Service) ReadLines(file string, startLine, count int) ([]string, error) {
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
	lines := make([]string, 0, count)
	for len(lines) < count && scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
