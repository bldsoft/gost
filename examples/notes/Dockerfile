# Build context is the gost repository root because the example lives inside the gost module:
#   docker build -f examples/notes/Dockerfile .
# A standalone service uses docs/templates/Dockerfile instead.
FROM golang:1.27-alpine AS builder

ARG COMMIT=unknown
ARG BRANCH=unknown

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -trimpath \
    -ldflags="-s -w \
      -X github.com/bldsoft/gost/version.Version=$(cat examples/notes/VERSION) \
      -X github.com/bldsoft/gost/version.GitCommit=${COMMIT} \
      -X github.com/bldsoft/gost/version.GitBranch=${BRANCH}" \
    -o /out/notes ./examples/notes/cmd/notes

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/notes /usr/local/bin/notes
USER 65532:65532
EXPOSE 3000
ENTRYPOINT ["/usr/local/bin/notes"]
