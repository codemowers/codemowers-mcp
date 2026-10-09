# Building Rust containers

Apply `building/general` for image stages, runtime identity and file permissions.

- Cache dependency work using Cargo manifests and the committed lockfile; include required workspace member manifests. Use `cargo fetch --locked` and `cargo build --release --locked --target <target>`. Do not update dependencies during image builds.
- If caching builds with placeholder sources, replace them with the complete real source tree and ensure Cargo rebuilds it. Exclude host `target` artifacts.
- Choose the target and linker for the deployment architecture. Native dependencies also need compatible libraries. Verify there is no dynamic interpreter or required shared library before choosing `scratch`.
- Copy required application assets. Use a compatible minimal runtime if static linking is unsuitable. Verify startup and TLS/native-library paths in the final image.

See `building/general` for the lolcatz Dockerfile examples; adapt the build and runtime stages to this language’s toolchain.
