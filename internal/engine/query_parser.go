package engine

import (
	internalquery "github.com/474420502/xjson/internal/query"
)

type Op = internalquery.Op

const (
	OpKey       = internalquery.OpKey
	OpIndex     = internalquery.OpIndex
	OpSlice     = internalquery.OpSlice
	OpFunc      = internalquery.OpFunc
	OpWildcard  = internalquery.OpWildcard
	OpRecursive = internalquery.OpRecursiveKey
	OpParent    = internalquery.OpParent
)

type queryToken struct {
	Op    Op
	Value interface{}
}

type slice struct {
	Start int
	End   int
}

const enableQueryCache = true

func ParseQuery(path string) ([]queryToken, error) {
	rawTokens, err := internalquery.NewParser(path).Parse()
	if err != nil {
		return nil, err
	}
	tokens := make([]queryToken, 0, len(rawTokens))
	for _, token := range rawTokens {
		adapted := queryToken{Op: Op(token.Type), Value: token.Value}
		if token.Type == internalquery.OpSlice {
			if rawSlice, ok := token.Value.([2]int); ok {
				adapted.Value = slice{Start: rawSlice[0], End: rawSlice[1]}
			}
		}
		tokens = append(tokens, adapted)
	}
	return tokens, nil
}