# Image URL to use all building/pushing image targets
IMG ?= digitalocean/clusterlint
TAG ?= dev

# Build the docker image
docker-build:
	docker build . -t ${IMG}:${TAG}

# Push the docker image
docker-push:
	docker push ${IMG}:${TAG}

# Build all binaries and sha sums for release
build-binaries:
	GOOS=linux GOARCH=amd64 go build -mod=vendor -o clusterlint ./cmd/clusterlint; tar -czvf clusterlint-${TAG}-linux-amd64.tar.gz ./clusterlint
	GOOS=linux GOARCH=386 go build -mod=vendor -o clusterlint ./cmd/clusterlint; tar -czvf clusterlint-${TAG}-linux-386.tar.gz ./clusterlint
	GOOS=darwin GOARCH=amd64 go build -mod=vendor -o clusterlint ./cmd/clusterlint; tar -czvf clusterlint-${TAG}-darwin-amd64.tar.gz ./clusterlint
	GOOS=darwin GOARCH=arm64 go build -mod=vendor -o clusterlint ./cmd/clusterlint; tar -czvf clusterlint-${TAG}-darwin-arm64.tar.gz ./clusterlint
	GOOS=windows GOARCH=amd64 go build -mod=vendor -o clusterlint.exe ./cmd/clusterlint; tar -czvf clusterlint-${TAG}-windows-amd64.tar.gz ./clusterlint
	GOOS=windows GOARCH=386 go build -mod=vendor -o clusterlint.exe ./cmd/clusterlint; tar -czvf clusterlint-${TAG}-windows-386.tar.gz ./clusterlint
	sha256sum clusterlint-${TAG}-linux-amd64.tar.gz >> clusterlint-${TAG}-checksums.sha256
	sha256sum clusterlint-${TAG}-linux-386.tar.gz >> clusterlint-${TAG}-checksums.sha256
	sha256sum clusterlint-${TAG}-darwin-amd64.tar.gz >> clusterlint-${TAG}-checksums.sha256
	sha256sum clusterlint-${TAG}-darwin-arm64.tar.gz >> clusterlint-${TAG}-checksums.sha256
	sha256sum clusterlint-${TAG}-windows-amd64.tar.gz >> clusterlint-${TAG}-checksums.sha256
	sha256sum clusterlint-${TAG}-windows-386.tar.gz >> clusterlint-${TAG}-checksums.sha256



