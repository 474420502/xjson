// Package scanner provides zero-allocation JSON scanning capabilities for xjson.
// It implements a high-performance scanner that can traverse JSON bytes directly
// without creating intermediate Go structures, similar to gjson's approach.
package scanner

// Scanner provides zero-allocation JSON scanning
type Scanner struct {
	data []byte
	pos  int
	len  int
}

// NewScanner creates a new scanner for the given JSON bytes
func NewScanner(data []byte) *Scanner {
	return &Scanner{
		data: data,
		pos:  0,
		len:  len(data),
	}
}

// Current returns the current byte
func (s *Scanner) Current() byte {
	if s.pos >= s.len {
		return 0
	}
	return s.data[s.pos]
}

// Peek returns the byte at the current position + offset without advancing
func (s *Scanner) Peek(offset int) byte {
	pos := s.pos + offset
	if pos >= s.len {
		return 0
	}
	return s.data[pos]
}

// Advance moves the position forward by n bytes
func (s *Scanner) Advance(n int) {
	s.pos += n
	if s.pos > s.len {
		s.pos = s.len
	}
}

// Reset resets the scanner position to the beginning
func (s *Scanner) Reset() {
	s.pos = 0
}

// SetPosition sets the scanner position
func (s *Scanner) SetPosition(pos int) {
	if pos < 0 {
		s.pos = 0
	} else if pos > s.len {
		s.pos = s.len
	} else {
		s.pos = pos
	}
}

// Position returns the current position
func (s *Scanner) Position() int {
	return s.pos
}

// Remaining returns the number of bytes remaining
func (s *Scanner) Remaining() int {
	return s.len - s.pos
}

// IsEOF returns true if the scanner is at the end
func (s *Scanner) IsEOF() bool {
	return s.pos >= s.len
}

// SkipWhitespace skips whitespace characters
func (s *Scanner) SkipWhitespace() {
	for s.pos < s.len {
		c := s.data[s.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			s.pos++
		} else {
			break
		}
	}
}

// ReadString reads a JSON string value and returns its content without quotes
func (s *Scanner) ReadString() (string, bool) {
	if s.pos >= s.len || s.data[s.pos] != '"' {
		return "", false
	}

	start := s.pos + 1 // Skip opening quote
	s.pos++

	for s.pos < s.len {
		c := s.data[s.pos]
		if c == '"' {
			// Found closing quote
			result := string(s.data[start:s.pos])
			s.pos++ // Skip closing quote
			return result, true
		} else if c == '\\' {
			// Handle escape sequences
			s.pos += 2 // Skip escape char and next char
		} else {
			s.pos++
		}
	}

	return "", false // Unterminated string
}

// ReadNumber reads a JSON number and returns its raw bytes
func (s *Scanner) ReadNumber() ([]byte, bool) {
	start := s.pos

	// Optional minus
	if s.pos < s.len && s.data[s.pos] == '-' {
		s.pos++
	}

	// Must have at least one digit
	if s.pos >= s.len || !isDigit(s.data[s.pos]) {
		s.pos = start
		return nil, false
	}

	// Read integer part
	if s.data[s.pos] == '0' {
		s.pos++
	} else {
		for s.pos < s.len && isDigit(s.data[s.pos]) {
			s.pos++
		}
	}

	// Optional fractional part
	if s.pos < s.len && s.data[s.pos] == '.' {
		s.pos++
		if s.pos >= s.len || !isDigit(s.data[s.pos]) {
			s.pos = start
			return nil, false
		}
		for s.pos < s.len && isDigit(s.data[s.pos]) {
			s.pos++
		}
	}

	// Optional exponent part
	if s.pos < s.len && (s.data[s.pos] == 'e' || s.data[s.pos] == 'E') {
		s.pos++
		if s.pos < s.len && (s.data[s.pos] == '+' || s.data[s.pos] == '-') {
			s.pos++
		}
		if s.pos >= s.len || !isDigit(s.data[s.pos]) {
			s.pos = start
			return nil, false
		}
		for s.pos < s.len && isDigit(s.data[s.pos]) {
			s.pos++
		}
	}

	return s.data[start:s.pos], true
}

// ReadBool reads a JSON boolean value
func (s *Scanner) ReadBool() (bool, bool) {
	if s.pos+4 <= s.len && string(s.data[s.pos:s.pos+4]) == "true" {
		s.pos += 4
		return true, true
	}
	if s.pos+5 <= s.len && string(s.data[s.pos:s.pos+5]) == "false" {
		s.pos += 5
		return false, true
	}
	return false, false
}

// ReadNull reads a JSON null value
func (s *Scanner) ReadNull() bool {
	if s.pos+4 <= s.len && string(s.data[s.pos:s.pos+4]) == "null" {
		s.pos += 4
		return true
	}
	return false
}

// FindKey searches for a specific key in the current object and positions the scanner
// at the start of its value. Returns true if found.
func (s *Scanner) FindKey(key string) bool {
	s.SkipWhitespace()

	if s.pos >= s.len || s.data[s.pos] != '{' {
		return false
	}

	s.pos++ // Skip opening brace
	s.SkipWhitespace()

	// Handle empty object
	if s.pos < s.len && s.data[s.pos] == '}' {
		return false
	}

	for {
		// Read key
		keyStr, ok := s.ReadString()
		if !ok {
			return false
		}

		s.SkipWhitespace()

		// Expect colon
		if s.pos >= s.len || s.data[s.pos] != ':' {
			return false
		}
		s.pos++ // Skip colon
		s.SkipWhitespace()

		if keyStr == key {
			return true // Found the key, scanner is positioned at value
		}

		// Skip the value
		if !s.SkipValue() {
			return false
		}

		s.SkipWhitespace()

		// Check for end of object
		if s.pos >= s.len {
			return false
		}

		if s.data[s.pos] == '}' {
			return false // End of object, key not found
		}

		if s.data[s.pos] == ',' {
			s.pos++ // Skip comma
			s.SkipWhitespace()
		} else {
			return false // Invalid JSON
		}
	}
}

// SkipValue skips over a JSON value at the current position
func (s *Scanner) SkipValue() bool {
	s.SkipWhitespace()

	if s.pos >= s.len {
		return false
	}

	c := s.data[s.pos]

	switch c {
	case '"':
		_, ok := s.ReadString()
		return ok
	case '{':
		return s.skipObject()
	case '[':
		return s.skipArray()
	case 't', 'f':
		_, ok := s.ReadBool()
		return ok
	case 'n':
		return s.ReadNull()
	default:
		if c == '-' || isDigit(c) {
			_, ok := s.ReadNumber()
			return ok
		}
		return false
	}
}

// skipObject skips over a JSON object
func (s *Scanner) skipObject() bool {
	if s.pos >= s.len || s.data[s.pos] != '{' {
		return false
	}

	s.pos++ // Skip opening brace
	depth := 1

	for s.pos < s.len && depth > 0 {
		c := s.data[s.pos]

		switch c {
		case '"':
			// Skip string
			s.pos++
			for s.pos < s.len {
				if s.data[s.pos] == '"' {
					s.pos++
					break
				} else if s.data[s.pos] == '\\' {
					s.pos += 2
				} else {
					s.pos++
				}
			}
		case '{':
			depth++
			s.pos++
		case '}':
			depth--
			s.pos++
		default:
			s.pos++
		}
	}

	return depth == 0
}

// skipArray skips over a JSON array
func (s *Scanner) skipArray() bool {
	if s.pos >= s.len || s.data[s.pos] != '[' {
		return false
	}

	s.pos++ // Skip opening bracket
	depth := 1

	for s.pos < s.len && depth > 0 {
		c := s.data[s.pos]

		switch c {
		case '"':
			// Skip string
			s.pos++
			for s.pos < s.len {
				if s.data[s.pos] == '"' {
					s.pos++
					break
				} else if s.data[s.pos] == '\\' {
					s.pos += 2
				} else {
					s.pos++
				}
			}
		case '[':
			depth++
			s.pos++
		case ']':
			depth--
			s.pos++
		default:
			s.pos++
		}
	}

	return depth == 0
}

// GetValueAt returns the raw bytes of the value at the current position
func (s *Scanner) GetValueAt() ([]byte, bool) {
	start := s.pos
	if s.SkipValue() {
		return s.data[start:s.pos], true
	}
	return nil, false
}

// Helper functions

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isHexDigit(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func isValidRuneInString(r rune) bool {
	return r != '"' && r != '\\' && r >= 0x20
}
