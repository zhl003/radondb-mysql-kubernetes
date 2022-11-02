FROM golang:1.17.13 as builder
WORKDIR /workspace
# Copy the go source
COPY cmd/mybackup/ cmd/mybackup/ 
COPY api api
COPY go.mod go.mod
COPY go.sum go.sum
ARG GO_PROXY=off
RUN if [ "$GO_PROXY" = "on" ]; then \
    go env -w GOPROXY=https://goproxy.cn,direct; \
    fi
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o mybackup cmd/mybackup/main.go 

FROM percona/percona-server:8.0.25 as mysql
FROM ubuntu:20.04
ARG MYSQLSH_VERSION=8.0.30-1
ARG DEBIAN_FRONTEND=noninteractive
COPY --from=builder /workspace/mybackup /usr/local/bin/mybackup
COPY --from=mysql /usr/bin/mysqlbinlog /usr/local/mysql/bin/mysqlbinlog
COPY --from=mysql /usr/bin/mysql       /usr/local/mysql/bin/mysql
RUN apt-get update \
  && apt-get install -y --no-install-recommends libjemalloc2 zstd python3 libpython3.8 s3cmd curl ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && curl -o /tmp/mysqlsh.deb -fsL https://dev.mysql.com/get/Downloads/MySQL-Shell/mysql-shell_${MYSQLSH_VERSION}ubuntu20.04_amd64.deb \
  && dpkg -i /tmp/mysqlsh.deb \
  && rm -f /tmp/mysqlsh.deb
ENV PATH=/usr/local/mysql/bin:"$PATH"
USER 10000:10000
ENTRYPOINT ["/mybackup"]

