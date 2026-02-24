package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

const crlf = "\r\n"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	// If CRLF is at the start of the data, it's the end of the headers
	if idx == 0 {
		return idx + 2, true, nil
	}

	headerLineString := string(data[:idx])
	parts := strings.Split(headerLineString, ": ")
	if len(parts) != 2 {
		return 0, false, fmt.Errorf("Malformed header: %s\n", headerLineString)
	}
	key := parts[0]
	for _, char := range key {
		if unicode.IsSpace(char) { // Checks if the rune is a whitespace character
			return 0, false, fmt.Errorf("Invalid header key: %s\n", key)
		}
	}
	value := parts[1]
	headervalue := strings.TrimSpace(value)
	if headervalue == "" {
		return 0, false, fmt.Errorf("Key contains only whitespaes")
	}

	h[key] = headervalue
	return idx + 2, false, nil
}
