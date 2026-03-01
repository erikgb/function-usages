package main

import (
	"context"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"

	v1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"

	"github.com/erikgb/function-usages/input/v1beta1"
)

func TestRunFunction(t *testing.T) {
	var (
		xr    = `{"apiVersion":"example.org/v1","kind":"XR","metadata":{"name":"cool-xr"},"spec":{"count":2}}`
		nxr   = `{"apiVersion":"example.org/v1","kind":"XR","metadata":{"name":"cool-xr","namespace":"cool-namespace"},"spec":{"count":2}}`
		nmr   = `{"apiVersion":"example.org/v1","kind":"MR","metadata":{"name":"cool-mr","namespace":"cool-namespace"}}`
		nuv2  = `{"apiVersion":"protection.crossplane.io/v1beta1","kind":"Usage","spec":{"by":{"apiVersion":"example.org/v1","kind":"MR","resourceRef":{"name":"cool-mr"}},"of":{"apiVersion":"example.org/v1","kind":"XR","resourceRef":{"name":"cool-xr"}},"reason":"dependency","replayDeletion":true}}`
		nu2v2 = `{"apiVersion":"protection.crossplane.io/v1beta1","kind":"Usage","spec":{"by":{"apiVersion":"example.org/v1","kind":"MR","resourceRef":{"name":"cool-mr"}},"of":{"apiVersion":"example.org/v1","kind":"MR","resourceRef":{"name":"cool-mr"}},"reason":"dependency","replayDeletion":true}}`
	)

	type args struct {
		ctx context.Context
		req *v1.RunFunctionRequest
	}
	type want struct {
		rsp *v1.RunFunctionResponse
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"FirstReadyUsageV2Namespaced": {
			reason: "The function should create a V2 Namespaced Usage when the first resource is ready",
			args: args{
				req: &v1.RunFunctionRequest{
					Input: resource.MustStructObject(&v1beta1.Input{
						Rules: []v1beta1.SequencingRule{
							{
								Sequence: []resource.Name{
									"first",
									"second",
								},
							},
						},
					}),
					Observed: &v1.State{
						Composite: &v1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
						Resources: map[string]*v1.Resource{
							"first": {
								Resource: resource.MustStructJSON(nxr),
							},
							"second": {
								Resource: resource.MustStructJSON(nmr),
							},
						},
					},
					Desired: &v1.State{
						Composite: &v1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
						Resources: map[string]*v1.Resource{
							"first": {
								Resource: resource.MustStructJSON(nxr),
							},
							"second": {
								Resource: resource.MustStructJSON(nmr),
							},
						},
					},
				},
			},
			want: want{
				rsp: &v1.RunFunctionResponse{
					Meta:    &v1.ResponseMeta{Ttl: durationpb.New(response.DefaultTTL)},
					Results: []*v1.Result{},
					Desired: &v1.State{
						Composite: &v1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
						Resources: map[string]*v1.Resource{
							"first": {
								Resource: resource.MustStructJSON(nxr),
							},
							"second": {
								Resource: resource.MustStructJSON(nmr),
							},
							"second-first-usage": {
								Resource: resource.MustStructJSON(nuv2),
								Ready:    v1.Ready_READY_TRUE,
							},
						},
					},
				},
			},
		},
		"MixedRegexUsageV2Namespaced": {
			reason: "The function should delay the creation of second and fourth resources because the first and third are not ready",
			args: args{
				req: &v1.RunFunctionRequest{
					Input: resource.MustStructObject(&v1beta1.Input{
						Rules: []v1beta1.SequencingRule{
							{
								Sequence: []resource.Name{
									"first",
									"second-.*",
									"third",
								},
							},
						},
					}),
					Observed: &v1.State{
						Composite: &v1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
						Resources: map[string]*v1.Resource{
							"first": {
								Resource: resource.MustStructJSON(nxr),
							},
							"second-0": {
								Resource: resource.MustStructJSON(nmr),
							},
							"second-1": {
								Resource: resource.MustStructJSON(nmr),
							},
							"third": {
								Resource: resource.MustStructJSON(nmr),
							},
						},
					},
					Desired: &v1.State{
						Composite: &v1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
						Resources: map[string]*v1.Resource{
							"first": {
								Resource: resource.MustStructJSON(nxr),
							},
							"second-0": {
								Resource: resource.MustStructJSON(nmr),
							},
							"second-1": {
								Resource: resource.MustStructJSON(nmr),
							},
							"third": {
								Resource: resource.MustStructJSON(nmr),
							},
						},
					},
				},
			},
			want: want{
				rsp: &v1.RunFunctionResponse{
					Meta:    &v1.ResponseMeta{Ttl: durationpb.New(response.DefaultTTL)},
					Results: []*v1.Result{},
					Desired: &v1.State{
						Composite: &v1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
						Resources: map[string]*v1.Resource{
							"first": {
								Resource: resource.MustStructJSON(nxr),
							},
							"second-0": {
								Resource: resource.MustStructJSON(nmr),
							},
							"second-1": {
								Resource: resource.MustStructJSON(nmr),
							},
							"second-0-first-usage": {
								Resource: resource.MustStructJSON(nuv2),
								Ready:    v1.Ready_READY_TRUE,
							},
							"second-1-first-usage": {
								Resource: resource.MustStructJSON(nuv2),
								Ready:    v1.Ready_READY_TRUE,
							},
							"third-second-0-usage": {
								Resource: resource.MustStructJSON(nu2v2),
								Ready:    v1.Ready_READY_TRUE,
							},
							"third-second-1-usage": {
								Resource: resource.MustStructJSON(nu2v2),
								Ready:    v1.Ready_READY_TRUE,
							},
							"third": {
								Resource: resource.MustStructJSON(nmr),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(tc.args.ctx, tc.args.req)

			if diff := cmp.Diff(tc.want.rsp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want rsp, +got rsp:\n%s", tc.reason, diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want err, +got err:\n%s", tc.reason, diff)
			}
		})
	}
}
