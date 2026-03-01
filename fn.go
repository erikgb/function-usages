package main

import (
	"context"
	"encoding/json"
	"maps"

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

	observedComposed, err := request.GetObservedComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get observed composed resources"))
		return rsp, nil
	}
	// Delete any usages from the observed composed resources
	maps.DeleteFunc(observedComposed, func(_ resource.Name, composed resource.ObservedComposed) bool {
		return isUsage(composed)
	})

	rules, err := compileRules(in.Rules)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "failed to compile rules"))
		return rsp, nil
	}

	g := graph.New(ResourceNameHash, graph.Directed(), graph.PreventCycles())
	for _, e := range buildEdges(rules, maps.Keys(observedComposed)) {
		for _, v := range []resource.Name{e.Source, e.Target} {
			if err := g.AddVertex(v); err != nil && !errors.Is(err, graph.ErrVertexAlreadyExists) {
				response.Fatal(rsp, errors.Wrapf(err, "unable to add vertex %s", v))
			}
		}

		if err := g.AddEdge(e.Source, e.Target); err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "unable to add edge %v", e))
		}
	}

	edges, err := g.Edges()
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get edges"))
		return rsp, nil
	}

	usages := make(map[resource.Name]*resource.DesiredComposed, len(edges))
	for _, e := range edges {
		of := e.Source
		by := e.Target

		f.log.Debug("Generate usage", "of", of, "by", by)
		usage := generateUsage(&observedComposed[of].Resource.Unstructured, &observedComposed[by].Resource.Unstructured)
		usageComposed := composed.New()
		if err := convertViaJSON(usageComposed, usage); err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot convert to JSON %s", usage))
			return rsp, err
		}
		f.log.Debug("created usage", "kind", usageComposed.GetKind())
		usages[by+"-"+of+"-usage"] = &resource.DesiredComposed{Resource: usageComposed, Ready: resource.ReadyTrue}
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

func generateUsage(of *unstructured.Unstructured, by *unstructured.Unstructured) map[string]any {
	return map[string]any{
		"apiVersion": protectionv1beta1.SchemeGroupVersion.String(),
		"kind":       protectionv1beta1.UsageKind,
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
			"reason":         "dependency",
			"replayDeletion": true,
		},
	}
}

func convertViaJSON(to, from any) error {
	bs, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, to)
}
