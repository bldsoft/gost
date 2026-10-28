# Multi-stage build for a gost service. Build from the repository root:
#   docker build --build-arg COMMIT=$(git rev-parse --short HEAD) --build-arg BRANCH=$(git rev-parse --abbrev-ref HEAD) -t <svc> .
FROM golang:1.27-alpine AS builder

ARG COMMIT=unknown
ARG BRANCH=unknown

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# version.Version/GitCommit/GitBranch feed /<svc>/version, discovery metadata and the log exporter.
RUN go build -trimpath \
    -ldflags="-s -w \
      -X github.com/bldsoft/gost/version.Version=$(cat VERSION) \
      -X github.com/bldsoft/gost/version.GitCommit=${COMMIT} \
      -X github.com/bldsoft/gost/version.GitBranch=${BRANCH}" \
    -o /out/<svc> ./cmd/<svc>

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/<svc> /usr/local/bin/<svc>
USER 65532:65532
EXPOSE 3000
ENTRYPOINT ["/usr/local/bin/<svc>"]
