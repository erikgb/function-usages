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
    reason: Available
    status: "True"
    type: Ready
  - lastTransitionTime: "2024-01-01T00:00:00Z"
    reason: Success
    status: "True"
    type: FunctionSuccess
---
apiVersion: render.crossplane.io/v1beta1
kind: Result
message: I was run with input [{[first-resource second-resource]} {[first-resource
  third-resource]}]!
severity: SEVERITY_NORMAL
step: deletion-ordering
```
