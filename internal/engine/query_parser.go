package engine

import (
	"sync"

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

type fastQuerySegment struct {
	key     string
	indices []int
}

type fastQueryPlan struct {
	segments []fastQuerySegment
}

const enableQueryCache = true

const maxCompiledQueryEntries = 512
const maxFastQueryPlanEntries = 512

var compiledQueryCache = struct {
	mu sync.RWMutex
	m  map[string][]queryToken
}{
	m: make(map[string][]queryToken),
}

var fastQueryPlanCache = struct {
	mu sync.RWMutex
	m  map[string]*fastQueryPlan
}{
	m: make(map[string]*fastQueryPlan),
}

func ParseQuery(path string) ([]queryToken, error) {
	if cached, ok := getCachedCompiledQuery(path); ok {
		return cached, nil
	}

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
	cacheCompiledQuery(path, tokens)
	return tokens, nil
}

func getCachedCompiledQuery(path string) ([]queryToken, bool) {
	compiledQueryCache.mu.RLock()
	defer compiledQueryCache.mu.RUnlock()
	tokens, ok := compiledQueryCache.m[path]
	return tokens, ok
}

func cacheCompiledQuery(path string, tokens []queryToken) {
	compiledQueryCache.mu.Lock()
	defer compiledQueryCache.mu.Unlock()
	if _, exists := compiledQueryCache.m[path]; exists {
		return
	}
	if len(compiledQueryCache.m) >= maxCompiledQueryEntries {
		return
	}
	compiledQueryCache.m[path] = tokens
}

func getFastQueryPlan(path string) (*fastQueryPlan, bool) {
	fastQueryPlanCache.mu.RLock()
	if plan, ok := fastQueryPlanCache.m[path]; ok {
		fastQueryPlanCache.mu.RUnlock()
		return plan, true
	}
	fastQueryPlanCache.mu.RUnlock()

	plan, ok := compileFastQueryPlan(path)
	if !ok {
		return nil, false
	}

	fastQueryPlanCache.mu.Lock()
	if existing, exists := fastQueryPlanCache.m[path]; exists {
		fastQueryPlanCache.mu.Unlock()
		return existing, true
	}
	if len(fastQueryPlanCache.m) < maxFastQueryPlanEntries {
		fastQueryPlanCache.m[path] = plan
	}
	fastQueryPlanCache.mu.Unlock()
	return plan, true
}

func compileFastQueryPlan(path string) (*fastQueryPlan, bool) {
	if path == "" || path == "/" || path == "//" {
		return &fastQueryPlan{}, true
	}
	if len(path) >= 2 && path[0] == '.' && path[1] == '.' {
		return nil, false
	}
	if len(path) >= 2 && path[0] == '/' && path[1] == '/' {
		return nil, false
	}

	i := 0
	for i < len(path) && path[i] == '/' {
		i++
	}
	if i >= len(path) {
		return &fastQueryPlan{}, true
	}

	segments := make([]fastQuerySegment, 0, 8)
	for i < len(path) {
		seg := fastQuerySegment{}
		kStart := i
		for i < len(path) && path[i] != '/' && path[i] != '[' {
			switch path[i] {
			case '*', '@', '.':
				return nil, false
			}
			i++
		}
		if kStart < i {
			seg.key = path[kStart:i]
		}

		for i < len(path) && path[i] == '[' {
			i++
			if i >= len(path) {
				return nil, false
			}
			neg := false
			if path[i] == '-' {
				neg = true
				i++
			}
			if i >= len(path) || path[i] < '0' || path[i] > '9' {
				return nil, false
			}
			value := 0
			for i < len(path) && path[i] >= '0' && path[i] <= '9' {
				value = value*10 + int(path[i]-'0')
				i++
			}
			if i >= len(path) || path[i] != ']' {
				return nil, false
			}
			if neg {
				value = -value
			}
			seg.indices = append(seg.indices, value)
			i++
		}

		if seg.key == "" && len(seg.indices) == 0 {
			return nil, false
		}
		segments = append(segments, seg)

		if i >= len(path) {
			break
		}
		if path[i] != '/' {
			return nil, false
		}
		if i+1 < len(path) && path[i+1] == '/' {
			return nil, false
		}
		for i < len(path) && path[i] == '/' {
			i++
		}
	}

	return &fastQueryPlan{segments: segments}, true
}