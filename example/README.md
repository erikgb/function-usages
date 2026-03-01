# Example manifests

You can run your function locally and test it using `crossplane render`
with these example manifests.

```shell
# Run the function locally
$ go run . --insecure --debug
```

```shell
# Then, in another terminal, call it with these example manifests
$ crossplane render xr.yaml composition.yaml functions.yaml -r
---
apiVersion: example.crossplane.io/v1
kind: XR
metadata:
  name: example-xr
status:
  conditions:
  - lastTransitionTime: "2024-01-01T00:00:00Z"
    message: 'Unready resources: first-subresource-1, first-subresource-2, second-object,
      and 1 more'
    reason: Creating
    status: "False"
    type: Ready
  - lastTransitionTime: "2024-01-01T00:00:00Z"
    reason: Success
    status: "True"
    type: FunctionSuccess
---
apiVersion: nop.crossplane.io/v1alpha1
kind: NopResource
metadata:
  annotations:
    crossplane.io/composition-resource-name: first-subresource-1
  labels:
    crossplane.io/composite: example-xr
  name: first-subresource-1
  ownerReferences:
  - apiVersion: example.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: XR
    name: example-xr
    uid: ""
---
apiVersion: nop.crossplane.io/v1alpha1
kind: NopResource
metadata:
  annotations:
    crossplane.io/composition-resource-name: first-subresource-2
  labels:
    crossplane.io/composite: example-xr
  name: first-subresource-2
  ownerReferences:
  - apiVersion: example.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: XR
    name: example-xr
    uid: ""
---
apiVersion: nop.crossplane.io/v1alpha1
kind: NopResource
metadata:
  annotations:
    crossplane.io/composition-resource-name: second-object
  labels:
    crossplane.io/composite: example-xr
  name: second-object
  ownerReferences:
  - apiVersion: example.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: XR
    name: example-xr
    uid: ""
---
apiVersion: nop.crossplane.io/v1alpha1
kind: NopResource
metadata:
  annotations:
    crossplane.io/composition-resource-name: third-resource
  labels:
    crossplane.io/composite: example-xr
  name: third-resource
  ownerReferences:
  - apiVersion: example.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: XR
    name: example-xr
    uid: ""
---
apiVersion: render.crossplane.io/v1beta1
kind: Result
message: I was run with input [{[first-subresource-.* object$ third-resource]}]!
severity: SEVERITY_NORMAL
step: deletion-ordering
```
