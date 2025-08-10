package query

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// Scanner represents a lexical scanner for the query language.
type Scanner struct {
	r *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// Scan returns the next token and literal value.
func (s *Scanner) Scan() Token {
	ch := s.read()

	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		s.unread()
		return s.scanIdent()
	}

	switch ch {
	case eof:
		return Token{Type: TokenEOF}
	case '/':
		if s.peek() == '/' {
			s.read()
			return Token{Type: TokenSlash2, Literal: "//"}
		}
		return Token{Type: TokenSlash, Literal: "/"}
	case '[':
		return Token{Type: TokenLBRACKET, Literal: "["}
	case ']':
		return Token{Type: TokenRBRACKET, Literal: "]"}
	case '@':
		return Token{Type: TokenCurrent, Literal: "@"}
	case '*':
		return Token{Type: TokenAll, Literal: "*"}
	case '.':
		if s.peek() == '.' {
			s.read()
			return Token{Type: TokenDOT2, Literal: ".."}
		}
		return Token{Type: TokenDOT, Literal: "."}
	case ':':
		return Token{Type: TokenColon, Literal: ":"}
	}

	return Token{Type: TokenIllegal, Literal: string(ch)}
}

func (s *Scanner) scanWhitespace() Token {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return Token{Type: TokenIllegal, Literal: buf.String()}
}

func (s *Scanner) scanIdent() Token {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) && ch != '_' {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	switch strings.ToUpper(buf.String()) {
	case "FILTER":
		return Token{Type: TokenFilter, Literal: buf.String()}
	}

	return Token{Type: TokenIdentity, Literal: buf.String()}
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *Scanner) peek() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	s.unread()
	return ch
}

func (s *Scanner) unread() { _ = s.r.UnreadRune() }

func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' || ch == '\n' }
func isLetter(ch rune) bool     { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }
func isDigit(ch rune) bool      { return (ch >= '0' && ch <= '9') }

var eof = rune(0)
