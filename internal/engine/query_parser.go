package engine

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/474420502/xjson/internal/core"
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

type fastQueryStepKind uint8

const (
	fastQueryStepKey fastQueryStepKind = iota + 1
	fastQueryStepIndex
)

type fastQueryStep struct {
	kind  fastQueryStepKind
	key   string
	index int
}

type specializedFastQueryKind uint8

const (
	specializedFastQueryNone specializedFastQueryKind = iota
	specializedFastQueryAllKeys
	specializedFastQueryKeysIndexKeys
)

type specializedFastQuery struct {
	kind         specializedFastQueryKind
	keys         []string
	suffixes     []string
	preKeys      []string
	preSuffixes  []string
	index        int
	arraySuffix  string
	postKeys     []string
	postSuffixes []string
}

type CompiledQuery struct {
	path        string
	fastPlan    *fastQueryPlan
	fastSteps   []fastQueryStep
	specialized *specializedFastQuery
	tokens      []queryToken
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
	tokens := adaptQueryTokens(rawTokens)
	cacheCompiledQuery(path, tokens)
	return tokens, nil
}

func CompileQuery(path string) (*CompiledQuery, error) {
	if plan, ok := compileFastQueryPlan(path); ok {
		steps := flattenFastQueryPlan(plan)
		return &CompiledQuery{path: path, fastPlan: plan, fastSteps: steps, specialized: buildSpecializedFastQuery(steps)}, nil
	}

	rawTokens, err := internalquery.NewParser(path).Parse()
	if err != nil {
		return nil, err
	}
	return &CompiledQuery{path: path, tokens: adaptQueryTokens(rawTokens)}, nil
}

func (cq *CompiledQuery) Path() string {
	if cq == nil {
		return ""
	}
	return cq.path
}

func (cq *CompiledQuery) Query(start core.Node) core.Node {
	if cq == nil {
		return newInvalidNode(fmt.Errorf("nil compiled query"))
	}
	if start == nil {
		return newInvalidNode(fmt.Errorf("nil start node"))
	}

	if enableQueryCache && cq.path != "" {
		if bn, ok := start.(interface {
			getCachedQueryResult(string) (core.Node, bool)
		}); ok {
			if cachedResult, exists := bn.getCachedQueryResult(cq.path); exists {
				return cachedResult
			}
		}
	}

	var result core.Node
	if cq.fastPlan != nil {
		result = executeFastQueryPlan(start, cq.fastPlan)
	} else if cq.specialized != nil {
		result = executeSpecializedFastQuery(start, cq.specialized)
	} else if len(cq.fastSteps) > 0 {
		result = executeFastQuerySteps(start, cq.fastSteps)
	} else {
		result = executeQueryTokens(start, cq.tokens)
	}

	if enableQueryCache && cq.path != "" {
		if bn, ok := start.(interface{ setCachedQueryResult(string, core.Node) }); ok {
			bn.setCachedQueryResult(cq.path, result)
		}
	}

	return result
}

func adaptQueryTokens(rawTokens []internalquery.QueryToken) []queryToken {
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
	return tokens
}

func flattenFastQueryPlan(plan *fastQueryPlan) []fastQueryStep {
	if plan == nil || len(plan.segments) == 0 {
		return nil
	}
	stepCount := 0
	for _, segment := range plan.segments {
		if segment.key != "" {
			stepCount++
		}
		stepCount += len(segment.indices)
	}
	if stepCount == 0 {
		return nil
	}
	steps := make([]fastQueryStep, 0, stepCount)
	for _, segment := range plan.segments {
		if segment.key != "" {
			steps = append(steps, fastQueryStep{kind: fastQueryStepKey, key: segment.key})
		}
		for _, idx := range segment.indices {
			steps = append(steps, fastQueryStep{kind: fastQueryStepIndex, index: idx})
		}
	}
	return steps
}

func buildSpecializedFastQuery(steps []fastQueryStep) *specializedFastQuery {
	if len(steps) == 0 {
		return nil
	}
	allKeys := true
	indexPos := -1
	for i, step := range steps {
		if step.kind != fastQueryStepKey {
			allKeys = false
		}
		if step.kind == fastQueryStepIndex {
			if indexPos != -1 {
				indexPos = -2
				break
			}
			indexPos = i
		}
	}
	if allKeys {
		keys := make([]string, 0, len(steps))
		for _, step := range steps {
			keys = append(keys, step.key)
		}
		return &specializedFastQuery{kind: specializedFastQueryAllKeys, keys: keys, suffixes: buildKeyChainSuffixes(keys)}
	}
	if indexPos >= 0 {
		for i, step := range steps {
			if i == indexPos {
				continue
			}
			if step.kind != fastQueryStepKey {
				return nil
			}
		}
		preKeys := make([]string, 0, indexPos)
		for _, step := range steps[:indexPos] {
			preKeys = append(preKeys, step.key)
		}
		postKeys := make([]string, 0, len(steps)-indexPos-1)
		for _, step := range steps[indexPos+1:] {
			postKeys = append(postKeys, step.key)
		}
		return &specializedFastQuery{
			kind:         specializedFastQueryKeysIndexKeys,
			preKeys:      preKeys,
			preSuffixes:  buildKeyIndexKeyPreSuffixes(preKeys, steps[indexPos].index, postKeys),
			index:        steps[indexPos].index,
			arraySuffix:  buildArraySuffix(steps[indexPos].index, postKeys),
			postKeys:     postKeys,
			postSuffixes: buildKeyChainSuffixes(postKeys),
		}
	}
	return nil
}

func buildKeyChainSuffixes(keys []string) []string {
	if len(keys) == 0 {
		return nil
	}
	suffixes := make([]string, len(keys))
	for i := range keys {
		suffixes[i] = buildSlashPath(keys[i:])
	}
	return suffixes
}

func buildKeyIndexKeyPreSuffixes(preKeys []string, index int, postKeys []string) []string {
	if len(preKeys) == 0 {
		return nil
	}
	suffixes := make([]string, len(preKeys))
	arraySuffix := buildArraySuffix(index, postKeys)
	for i := range preKeys {
		prefix := buildSlashPath(preKeys[i:])
		suffixes[i] = prefix + arraySuffix
	}
	return suffixes
}

func buildArraySuffix(index int, postKeys []string) string {
	var builder strings.Builder
	builder.Grow(8 + len(postKeys)*8)
	builder.WriteByte('[')
	builder.WriteString(strconv.Itoa(index))
	builder.WriteByte(']')
	if len(postKeys) > 0 {
		builder.WriteString(buildSlashPath(postKeys))
	}
	return builder.String()
}

func buildSlashPath(keys []string) string {
	if len(keys) == 0 {
		return ""
	}
	var builder strings.Builder
	total := 0
	for _, key := range keys {
		total += len(key) + 1
	}
	builder.Grow(total)
	for _, key := range keys {
		builder.WriteByte('/')
		builder.WriteString(key)
	}
	return builder.String()
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
