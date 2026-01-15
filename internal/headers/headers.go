package headers

import (
	"bytes"
	"fmt"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if len(data) == 0 {
		return 0, false, nil
	}

	parts := bytes.Split(data, []byte(crlf))

	if len(parts) >= 1 && len(parts[0]) == 0 {
		return 2, true, nil
	}

	if len(parts) >= 2 && len(parts[0]) == 0 {
		return 2, false, nil
	}

	firstHeaderBytes := 0

	for i, part := range parts {

		if len(part) == 0 {
			if i == 0 {
				return 2, true, nil
			}
			if firstHeaderBytes == 0 {
				return 0, false, nil
			}
			return firstHeaderBytes, false, nil
		}

		// If this is the last "part" and the data doesn't end with CRLF,
		// then this header line is incomplete.
		if i == len(parts)-1 && !bytes.HasSuffix(data, []byte(crlf)) {
			if firstHeaderBytes == 0 {
				return 0, false, nil
			}
			return firstHeaderBytes, false, nil
		}

		headerParts := bytes.SplitN(part, []byte(":"), 2)
		if len(headerParts) != 2 {
			return 0, false, fmt.Errorf("invalid header line")
		}

		key := strings.ToLower(string(headerParts[0]))

		if key != strings.TrimRight(key, " ") {
			return 0, false, fmt.Errorf("invalid header name: %s", key)
		}

		value := bytes.TrimSpace(headerParts[1])
		key = strings.TrimSpace(key)
		if !validTokens([]byte(key)) {
			return 0, false, fmt.Errorf("invalid header token found: %s", key)
		}

		h.Set(key, string(value))
		headerBytes := len(part) + len(crlf)

		if firstHeaderBytes == 0 {
			firstHeaderBytes = headerBytes
		}

		// If next part is empty, we've reached the end-of-headers marker,
		// but still only report the first header's bytes as consumed.
		if i+1 < len(parts) && len(parts[i+1]) == 0 {
			return firstHeaderBytes, false, nil
		}
	}

	return firstHeaderBytes, false, nil
}

func (h Headers) Set(key, value string) {

	key = strings.ToLower(key)
	v, ok := h[key]
	if ok {
		value = strings.Join([]string{
			v,
			value,
		}, ", ")
	}
	h[key] = value
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Replace(key string, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) Delete(key string) {
	delete(h, strings.ToLower(key))
}

var tokenChars = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

// validTokens checks if the data contains only valid tokens
// or characters that are allowed in a token
func validTokens(data []byte) bool {
	for _, c := range data {
		if !(c >= 'A' && c <= 'Z' ||
			c >= 'a' && c <= 'z' ||
			c >= '0' && c <= '9' ||
			c == '-') {
			return false
		}
	}
	return true
}
