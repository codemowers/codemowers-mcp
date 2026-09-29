# Building Go containers

Apply `building/general` for image and runtime conventions.

- Copy `go.mod` and `go.sum` before sources; run `go mod download` to cache dependency downloads separately.
- Use `CGO_ENABLED=0` only when dependencies support it. CGo applications need compatible runtime libraries or verified static linking. Optional `-ldflags="-s -w"` removes symbol/debug information.
- Copy assets to their runtime paths. A `scratch` image also needs any CA bundle or other runtime data used by the application; verify the packaged image starts and serves required assets.

See the [lolcatz Go service Dockerfile](https://github.com/codemowers/lolcatz/blob/main/services/browse/Dockerfile) for a complete example.
