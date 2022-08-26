#
# Step 1 : build executable binary
#
FROM --platform=$BUILDPLATFORM golang:latest AS builder
ARG TARGETOS
ARG TARGETARCH

COPY / /host-files
RUN go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@master

WORKDIR /host-files

#without CGO_ENABLED=0 the target environment needs glibc. Setting to zero does link statically
RUN go generate ./...
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH  go build ./cmd/server

#
# Step 2 : build small image
#
FROM --platform=$TARGETPLATFORM alpine:latest

COPY --from=builder /host-files/server /server
#setup timezone
ENV TZ=Europe/Berlin
ENV ZONEINFO=/zoneinfo.zip
ENTRYPOINT ["/server"]

