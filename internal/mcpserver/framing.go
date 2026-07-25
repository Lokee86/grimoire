package mcpserver

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var errMessageTooLarge = errors.New("MCP message exceeds configured limit")

type messageReader struct {
	reader *bufio.Reader
	limit  int
}

func newMessageReader(reader io.Reader, limit int) *messageReader {
	return &messageReader{reader: bufio.NewReader(reader), limit: limit}
}

func (reader *messageReader) next() ([]byte, error) {
	for {
		line, err := reader.readLine()
		if err != nil {
			return nil, err
		}
		trimmed := strings.TrimSpace(string(line))
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(trimmed), "content-length:") {
			return reader.readFramed(trimmed)
		}
		if len(line) > reader.limit {
			return nil, errMessageTooLarge
		}
		return line, nil
	}
}

func (reader *messageReader) readFramed(firstHeader string) ([]byte, error) {
	length, err := parseContentLength(firstHeader)
	if err != nil {
		return nil, err
	}
	for {
		line, readErr := reader.readLine()
		if readErr != nil {
			return nil, readErr
		}
		if len(strings.TrimSpace(string(line))) == 0 {
			break
		}
	}
	if length < 0 || length > reader.limit {
		return nil, errMessageTooLarge
	}
	body := make([]byte, length)
	_, err = io.ReadFull(reader.reader, body)
	return body, err
}

func (reader *messageReader) readLine() ([]byte, error) {
	line, err := reader.reader.ReadBytes('\n')
	if len(line) > reader.limit {
		return nil, errMessageTooLarge
	}
	if err != nil && len(line) == 0 {
		return nil, err
	}
	return []byte(strings.TrimRight(string(line), "\r\n")), nil
}

func parseContentLength(header string) (int, error) {
	_, value, found := strings.Cut(header, ":")
	if !found {
		return 0, fmt.Errorf("invalid Content-Length header")
	}
	length, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || length < 0 {
		return 0, fmt.Errorf("invalid Content-Length header")
	}
	return length, nil
}
