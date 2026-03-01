package main

import (
	"fmt"
	"iter"
	"regexp"
	"strings"

	"github.com/dominikbraun/graph"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/resource"

	"github.com/erikgb/function-usages/input/v1beta1"
)

type compiledRule struct {
	Patterns []*regexp.Regexp
}

func compileRules(rules []v1beta1.SequencingRule) ([]compiledRule, error) {
	var compiled []compiledRule

	for _, rule := range rules {
		cr := compiledRule{}
		for _, pattern := range rule.Sequence {
			re, err := getStrictRegex(string(pattern))
			if err != nil {
				return nil, errors.Wrapf(err, "cannot compile regex %s", pattern)
			}
			cr.Patterns = append(cr.Patterns, re)
		}
		compiled = append(compiled, cr)
	}

	return compiled, nil
}

func getStrictRegex(pattern string) (*regexp.Regexp, error) {
	const (
		// START marks the start of a regex pattern.
		START = "^"
		// END marks the end of a regex pattern.
		END = "$"
	)

	if !strings.HasPrefix(pattern, START) {
		pattern = START + pattern
	}
	if !strings.HasSuffix(pattern, END) {
		pattern += END
	}
	return regexp.Compile(pattern)
}

func buildEdges(rules []compiledRule, names iter.Seq[resource.Name]) []graph.Edge[resource.Name] {
	edgesMap := make(map[string]graph.Edge[resource.Name]) // key = "from->to"

	for _, rule := range rules {
		for i := range len(rule.Patterns) - 1 {
			fromRe := rule.Patterns[i]
			toRe := rule.Patterns[i+1]

			var fromMatches, toMatches []resource.Name
			for r := range names {
				if fromRe.MatchString(string(r)) {
					fromMatches = append(fromMatches, r)
				}
				if toRe.MatchString(string(r)) {
					toMatches = append(toMatches, r)
				}
			}

			for _, from := range fromMatches {
				for _, to := range toMatches {
					key := fmt.Sprintf("%s->%s", from, to)
					if _, exists := edgesMap[key]; !exists {
						edgesMap[key] = graph.Edge[resource.Name]{Source: from, Target: to}
					}
				}
			}
		}
	}

	edges := make([]graph.Edge[resource.Name], 0, len(edgesMap))
	for _, e := range edgesMap {
		edges = append(edges, e)
	}
	return edges
}

// ResourceNameHash is a hashing function that accepts a resource name and uses the equivalent
// string as a hash value. Using it as Hash will yield a Graph[resource.Name, string].
func ResourceNameHash(v resource.Name) resource.Name {
	return v
}
