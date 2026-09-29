FROM golang:1.26.5-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY *.go ./
COPY content ./content
COPY web ./web
ARG BUILD_REVISION=dev
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -buildvcs=false -trimpath -ldflags="-s -w -X main.version=${BUILD_REVISION}" -o /codemowers-mcp .

FROM scratch
COPY --from=build /codemowers-mcp /codemowers-mcp
EXPOSE 3000 3001
ENTRYPOINT ["/codemowers-mcp"]
