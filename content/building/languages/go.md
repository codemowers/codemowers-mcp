# Building Go containers

Apply `building/general` for image and runtime conventions.

- Copy `go.mod` and `go.sum` before sources; run `go mod download` to cache dependency downloads separately.
- Use `CGO_ENABLED=0` only when dependencies support it. CGo applications need compatible runtime libraries or verified static linking. Optional `-ldflags="-s -w"` removes symbol/debug information.
- Copy required application assets to their runtime paths; verify the packaged image starts and serves them. Runtime CA trust follows `building/general`, including for `scratch` images.

See the [lolcatz Go service Dockerfile](https://github.com/codemowers/lolcatz/blob/main/services/browse/Dockerfile) for a complete example.
