# Build the manager binary
FROM golang:1.17.13 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
ARG GO_PROXY=off
RUN if [ "$GO_PROXY" = "on" ]; then \
    go env -w GOPROXY=https://goproxy.cn,direct; \
    fi
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN  go mod download

# Copy the go source
COPY cmd/manager/main.go cmd/manager/main.go
COPY api/ api/
COPY mysqlcluster/ mysqlcluster/
COPY controllers/ controllers/
COPY backup/ backup/
COPY internal/ internal/
COPY utils/ utils/
COPY mysqluser/ mysqluser/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager cmd/manager/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM radondb/distroless:nonroot
# If you make a image manually, the cn environment may not be able to access .io, you can switch to the following path
# FROM radondb/distroless:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
