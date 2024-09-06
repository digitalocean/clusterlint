# Build the clusterlint binary
FROM golang:1.23 as builder
WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/

# Copy the go source
COPY cmd/clusterlint/main.go main.go
COPY kube/ kube/
COPY checks checks/

# Build
ARG version
RUN GOFLAGS="-mod=vendor" CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
    go build -ldflags="-X main.Version=$version" -a -o clusterlint main.go

# Use distroless as minimal base image to package the clusterlint binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/clusterlint .
USER nonroot:nonroot

ENTRYPOINT ["/clusterlint"]
CMD ["-h"]
