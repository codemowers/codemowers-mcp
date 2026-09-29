# Building Python containers

Apply `building/general` for locked dependencies, runtime libraries and image contents. Install into the container's Python environment without an extra virtual environment; with uv, set `UV_PROJECT_ENVIRONMENT=/usr/local`.

Virtual environments remain appropriate outside containers. Agent development workflow follows `developing/local-development`.

See the [lolcatz Python OCR service Dockerfile](https://github.com/codemowers/lolcatz/blob/main/services/ocr/Dockerfile) for a complete example.
