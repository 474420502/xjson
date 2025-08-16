package engine

import (
	"unsafe"
)

// FastCharSearch uses SIMD-like optimizations for character searching
// This implementation uses word-level scanning to find characters faster

const wordSize = 8 // 64-bit word size

// findQuoteOptimized uses adaptive character scanning based on data size
func findQuoteOptimized(data []byte, start int) int {
	if start >= len(data) || data[start] != '"' {
		return -1
	}

	remaining := len(data) - start - 1

	// For small data, use simple byte-by-byte scanning
	if remaining < 64 {
		return findQuoteSimple(data, start)
	}

	// For larger data, use word-level scanning
	return findQuoteWord(data, start)
}

// findQuoteSimple is the original simple implementation for small data
func findQuoteSimple(data []byte, start int) int {
	i := start + 1
	for i < len(data) {
		if data[i] == '"' {
			// Check if escaped
			escapes := 0
			j := i - 1
			for j >= start+1 && data[j] == '\\' {
				escapes++
				j--
			}
			if escapes%2 == 0 {
				return i
			}
		}
		i++
	}
	return -1
}

// findQuoteWord uses word-level scanning for larger data
func findQuoteWord(data []byte, start int) int {
	i := start + 1
	length := len(data)

	// Fast path: scan 8 bytes at a time when possible
	for i+wordSize <= length {
		// Load 8 bytes as uint64
		word := *(*uint64)(unsafe.Pointer(&data[i]))

		// Check for quote or backslash in parallel using bit manipulation
		// This technique checks all 8 bytes simultaneously for the target characters

		// Create masks for quotes (0x22) and backslashes (0x5C)
		quoteMask := word ^ 0x2222222222222222     // XOR with repeated quote bytes
		backslashMask := word ^ 0x5C5C5C5C5C5C5C5C // XOR with repeated backslash bytes

		// Use bit manipulation to detect zeros (indicating matches)
		// This checks if any byte in the word equals our target characters
		quoteFound := hasZeroByte(quoteMask)
		backslashFound := hasZeroByte(backslashMask)

		if quoteFound || backslashFound {
			// Found a potential match, scan byte by byte in this word
			for j := 0; j < wordSize && i+j < length; j++ {
				c := data[i+j]
				if c == '"' {
					// Check if escaped
					escapes := 0
					k := i + j - 1
					for k >= start+1 && data[k] == '\\' {
						escapes++
						k--
					}
					if escapes%2 == 0 {
						return i + j
					}
				}
			}
		}
		i += wordSize
	}

	// Handle remaining bytes
	for i < length {
		if data[i] == '"' {
			// Check if escaped
			escapes := 0
			j := i - 1
			for j >= start+1 && data[j] == '\\' {
				escapes++
				j--
			}
			if escapes%2 == 0 {
				return i
			}
		}
		i++
	}

	return -1
}

// findBraceOptimized uses optimized scanning for braces with nesting
func findBraceOptimized(data []byte, start int) int {
	if start >= len(data) || data[start] != '{' {
		return -1
	}

	level := 1
	i := start + 1
	length := len(data)
	inString := false

	// Fast scanning when not in string
	for i < length && level > 0 {
		if !inString {
			// Use word-level scanning when not in string for better performance
			for i+wordSize <= length && level > 0 && !inString {
				word := *(*uint64)(unsafe.Pointer(&data[i]))

				// Create masks for special characters: {, }, "
				braceMask := word ^ 0x7B7B7B7B7B7B7B7B // '{'
				closeMask := word ^ 0x7D7D7D7D7D7D7D7D // '}'
				quoteMask := word ^ 0x2222222222222222 // '"'

				if hasZeroByte(braceMask) || hasZeroByte(closeMask) || hasZeroByte(quoteMask) {
					// Found special character, process byte by byte
					break
				}
				i += wordSize
			}

			// Handle individual bytes
			if i < length {
				c := data[i]
				switch c {
				case '"':
					inString = true
				case '{':
					level++
				case '}':
					level--
				}
				i++
			}
		} else {
			// In string, look for closing quote
			for i < length && inString {
				if data[i] == '"' {
					// Check if escaped
					escapes := 0
					j := i - 1
					for j >= 0 && data[j] == '\\' {
						escapes++
						j--
					}
					if escapes%2 == 0 {
						inString = false
					}
				}
				i++
			}
		}
	}

	if level == 0 {
		return i - 1
	}
	return -1
}

// findBracketOptimized uses optimized scanning for array brackets
func findBracketOptimized(data []byte, start int) int {
	if start >= len(data) || data[start] != '[' {
		return -1
	}

	level := 1
	i := start + 1
	length := len(data)
	inString := false

	for i < length && level > 0 {
		if !inString {
			// Use word-level scanning for better performance
			for i+wordSize <= length && level > 0 && !inString {
				word := *(*uint64)(unsafe.Pointer(&data[i]))

				// Create masks for special characters: [, ], "
				openMask := word ^ 0x5B5B5B5B5B5B5B5B  // '['
				closeMask := word ^ 0x5D5D5D5D5D5D5D5D // ']'
				quoteMask := word ^ 0x2222222222222222 // '"'

				if hasZeroByte(openMask) || hasZeroByte(closeMask) || hasZeroByte(quoteMask) {
					// Found special character, process byte by byte
					break
				}
				i += wordSize
			}

			// Handle individual bytes
			if i < length {
				c := data[i]
				switch c {
				case '"':
					inString = true
				case '[':
					level++
				case ']':
					level--
				}
				i++
			}
		} else {
			// In string, look for closing quote with escape handling
			for i < length && inString {
				if data[i] == '"' {
					// Check if escaped
					escapes := 0
					j := i - 1
					for j >= 0 && data[j] == '\\' {
						escapes++
						j--
					}
					if escapes%2 == 0 {
						inString = false
					}
				}
				i++
			}
		}
	}

	if level == 0 {
		return i - 1
	}
	return -1
}

// hasZeroByte checks if a 64-bit word contains any zero bytes
// This is a classic bit manipulation technique for SIMD-like operations
func hasZeroByte(word uint64) bool {
	// This technique uses the fact that (x - 0x0101010101010101) & ^x & 0x8080808080808080
	// will be non-zero if and only if x contains at least one zero byte
	return (word-0x0101010101010101)&^word&0x8080808080808080 != 0
}
