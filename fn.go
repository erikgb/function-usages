package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"regexp"
	"strings"

	protectionv1beta1 "github.com/crossplane/crossplane/v2/apis/protection/v1beta1"
	"github.com/dominikbraun/graph"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"

	"github.com/erikgb/function-usages/input/v1beta1"
)

// Function returns whatever response you ask it to.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	in := &v1beta1.Input{}
	if err := request.GetInput(req, in); err != nil {
		// You can set a custom status condition on the claim. This allows you to
		// communicate with the user. See the link below for status condition
		// guidance.
		// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
		response.ConditionFalse(rsp, "FunctionSuccess", "InternalError").
			WithMessage("Something went wrong.").
			TargetCompositeAndClaim()

		// You can emit an event regarding the claim. This allows you to communicate
		// with the user. Note that events should be used sparingly and are subject
		// to throttling; see the issue below for more information.
		// https://github.com/crossplane/crossplane/issues/5802
		response.Warning(rsp, errors.New("something went wrong")).
			TargetCompositeAndClaim()

		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from %T", req))
		return rsp, nil
	}

	f.log.Info("I was run!", "input", in.Rules)

	rules, err := compileRules(in.Rules)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot compile rules"))
		return rsp, nil
	}

	observedComposed, err := request.GetObservedComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get observed composed resources"))
		return rsp, nil
	}
	// Delete any usages from the observed composed resources
	maps.DeleteFunc(observedComposed, func(_ resource.Name, composed resource.ObservedComposed) bool {
		return isUsage(composed)
	})

	g := graph.New(ResourceNameHash, graph.Directed(), graph.PreventCycles())

	for _, e := range buildEdges(rules, maps.Keys(observedComposed)) {
		for _, v := range []resource.Name{e.From, e.To} {
			if err := g.AddVertex(v); err != nil && !errors.Is(err, graph.ErrVertexAlreadyExists) {
				response.Fatal(rsp, errors.Wrapf(err, "unable to add vertex %s", v))
			}
		}

		if err := g.AddEdge(e.From, e.To); err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "unable to add edge %v", e))
		}
	}

	usages := make(map[resource.Name]*resource.DesiredComposed)

	edges, err := g.Edges()
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get edges"))
		return rsp, nil
	}
	for _, e := range edges {
		f.log.Debug("Generate Usage of ", "k:", e.Target, "by c:", e.Source)
		usage := GenerateV2Usage(&observedComposed[e.Source].Resource.Unstructured, &observedComposed[e.Target].Resource.Unstructured)
		usageComposed := composed.New()
		if err := convertViaJSON(usageComposed, usage); err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot convert to JSON %s", usage))
			return rsp, err
		}
		f.log.Debug("created usage", "kind", usageComposed.GetKind(), "name", usageComposed.GetName(), "namespace", usageComposed.GetNamespace())
		usages[e.Target+"-"+e.Source+"-usage"] = &resource.DesiredComposed{Resource: usageComposed, Ready: resource.ReadyTrue}
	}

	desiredComposed, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get desired composed resources"))
		return rsp, nil
	}

	maps.Copy(desiredComposed, usages)
	rsp.Desired.Resources = nil
	return rsp, response.SetDesiredComposedResources(rsp, desiredComposed)
}

func isUsage(composed resource.ObservedComposed) bool {
	kind := composed.Resource.GetKind()
	if kind != protectionv1beta1.UsageKind {
		return false
	}

	gv, _ := schema.ParseGroupVersion(composed.Resource.GetAPIVersion())
	return gv.Group == protectionv1beta1.Group
}

type edge struct {
	From resource.Name
	To   resource.Name
}

func buildEdges(rules []compiledRule, observedComposed iter.Seq[resource.Name]) []edge {
	edgesMap := make(map[string]edge) // key = "from->to"

	for _, rule := range rules {
		for i := range len(rule.Patterns) - 1 {
			fromRe := rule.Patterns[i]
			toRe := rule.Patterns[i+1]

			var fromMatches, toMatches []resource.Name
			for r := range observedComposed {
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
						edgesMap[key] = edge{From: from, To: to}
					}
				}
			}
		}
	}

	edges := make([]edge, 0, len(edgesMap))
	for _, e := range edgesMap {
		edges = append(edges, e)
	}
	return edges
}

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

// ResourceNameHash is a hashing function that accepts a resource name and uses the equivalent
// string as a hash value. Using it as Hash will yield a Graph[resource.Name, string].
func ResourceNameHash(v resource.Name) resource.Name {
	return v
}

// GenerateV2Usage creates a v2 Usage for a resource.
func GenerateV2Usage(of *unstructured.Unstructured, by *unstructured.Unstructured) map[string]any {
	const (
		DependencyReason = "dependency"
		UsageNameSuffix  = "dependency"
	)

	name := strings.ToLower(by.GetKind() + "-" + by.GetName() + "-" + of.GetKind() + "-" + of.GetName())
	usageType := protectionv1beta1.UsageKind
	usageMeta := map[string]any{
		"namespace": of.GetNamespace(),
		"name":      GenerateName(name, UsageNameSuffix),
	}

	usage := map[string]any{
		"apiVersion": protectionv1beta1.SchemeGroupVersion.String(),
		"kind":       usageType,
		"metadata":   usageMeta,
		"spec": map[string]any{
			"by": map[string]any{
				"apiVersion": by.GetAPIVersion(),
				"kind":       by.GetKind(),
				"resourceRef": map[string]any{
					"name": by.GetName(),
				},
			},
			"of": map[string]any{
				"apiVersion": of.GetAPIVersion(),
				"kind":       of.GetKind(),
				"resourceRef": map[string]any{
					"name": of.GetName(),
				},
			},
			"reason":         DependencyReason,
			"replayDeletion": true,
		},
	}
	return usage
}

// GenerateName generates a valid Kubernetes name.
func GenerateName(name, suffix string) string {
	const (
		// maxKubernetesNameLength is the maximum length allowed for Kubernetes resource names.
		maxKubernetesNameLength = 63
		// hashLength is the length of the hash to apply to names.
		hashLength = 6
	)

	h := sha256.Sum256([]byte(name))
	hEncoded := hex.EncodeToString(h[:])[:hashLength]
	fullSuffix := hEncoded + "-" + suffix
	fullName := name + "-" + fullSuffix

	if len(fullName) <= maxKubernetesNameLength {
		return fullName
	}

	maxNameLength := maxKubernetesNameLength - len(fullSuffix) - 1 // -1 for the hyphen separator
	truncatedName := name[:maxNameLength]

	// Ensure the truncated name ends with a hyphen
	if !strings.HasSuffix(truncatedName, "-") {
		truncatedName += "-"
	}

	return truncatedName + fullSuffix
}

func convertViaJSON(to, from any) error {
	bs, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, to)
}
