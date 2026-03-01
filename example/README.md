# Example manifests

You can run your function locally and test it using `crossplane render`
with these example manifests.

```shell
# Run the function locally
$ go run . --insecure --debug
```

```shell
# Then, in another terminal, call it with these example manifests
$ crossplane render xr.yaml composition.yaml functions.yaml -r -o observed.yaml
---
apiVersion: example.crossplane.io/v1
kind: XR
metadata:
  name: example-xr
status:
  conditions:
  - lastTransitionTime: "2024-01-01T00:00:00Z"
    message: 'Unready resources: first-subresource-1, first-subresource-2, second-resource,
      and 1 more'
    reason: Creating
    status: "False"
    type: Ready
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
    crossplane.io/composition-resource-name: second-resource
  labels:
    crossplane.io/composite: example-xr
  name: second-resource
  ownerReferences:
  - apiVersion: example.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: XR
    name: example-xr
    uid: ""
---
apiVersion: protection.crossplane.io/v1beta1
kind: Usage
metadata:
  annotations:
    crossplane.io/composition-resource-name: second-resource-first-subresource-1-usage
  generateName: example-xr-
  labels:
    crossplane.io/composite: example-xr
  ownerReferences:
  - apiVersion: example.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: XR
    name: example-xr
    uid: ""
spec:
  by:
    apiVersion: nop.crossplane.io/v1alpha1
    kind: NopResource
    resourceRef:
      name: second-resource
  of:
    apiVersion: nop.crossplane.io/v1alpha1
    kind: NopResource
    resourceRef:
      name: first-subresource-1
  reason: dependency
  replayDeletion: true
---
apiVersion: protection.crossplane.io/v1beta1
kind: Usage
metadata:
  annotations:
    crossplane.io/composition-resource-name: second-resource-first-subresource-2-usage
  generateName: example-xr-
  labels:
    crossplane.io/composite: example-xr
  ownerReferences:
  - apiVersion: example.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: XR
    name: example-xr
    uid: ""
spec:
  by:
    apiVersion: nop.crossplane.io/v1alpha1
    kind: NopResource
    resourceRef:
      name: second-resource
  of:
    apiVersion: nop.crossplane.io/v1alpha1
    kind: NopResource
    resourceRef:
      name: first-subresource-2
  reason: dependency
  replayDeletion: true
---
apiVersion: nop.crossplane.io/v1alpha1
kind: NopResource
metadata:
  annotations:
    crossplane.io/composition-resource-name: third-resource
  generateName: example-xr-
  labels:
    crossplane.io/composite: example-xr
  ownerReferences:
  - apiVersion: example.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: XR
    name: example-xr
    uid: ""
```
