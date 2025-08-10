package query

// Op represents a query operation type.
type Op int

const (
	OpKey Op = iota
	OpIndex
	OpSlice
	OpFunc
	OpWildcard
	OpRecursiveKey
	OpParent
	OpAll
)

// QueryToken represents a single token in a parsed query.
type QueryToken struct {
	Type  Op
	Value interface{}
}

// TokenType represents the type of a lexical token.
type TokenType int

// Token represents a lexical token with its type, literal value, and position.
type Token struct {
	Type    TokenType
	Literal string
	Pos     int
}

const (
	TokenIllegal TokenType = iota
	TokenEOF
	TokenIdentity
	TokenSlash
	TokenSlash2
	TokenLBRACKET
	TokenRBRACKET
	TokenCurrent
	TokenAll
	TokenDOT
	TokenDOT2
	TokenFilter
	TokenColon
)
