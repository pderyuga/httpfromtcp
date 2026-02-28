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
	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	key = strings.TrimSpace(key)

	for _, char := range key {
		if !isTchar(char) { // Checks if the rune is a whitespace character
			return 0, false, fmt.Errorf("Invalid header name: %s\n", key)
		}
	}
	value := parts[1]
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false, fmt.Errorf("Header contains only whitespaes")
	}

	h.Set(key, value)
	return idx + 2, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	existingValue, ok := h[key]
	if ok {
		h[key] = existingValue + ", " + value
	} else {
		h[key] = value
	}
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	value, ok := h[key]
	return value, ok
}

func (h Headers) Override(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) Remove(key string) {
	key = strings.ToLower(key)
	delete(h, key)
}

func isTchar(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsNumber(r) {
		return true
	}
	// Define the set of allowed symbols: !#$%&'*+-.^_`|~
	switch r {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '_', '`', '|', '~':
		return true
	}
	return false
}
