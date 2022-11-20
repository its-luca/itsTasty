#
# Step 1 : build golang backend
#
FROM --platform=$BUILDPLATFORM golang:latest AS backend-builder
ARG TARGETOS
ARG TARGETARCH


WORKDIR /host-files

#install deps first, to allow caching installation if mod and sum file did not change
COPY /go.mod .
COPY /go.sum .
RUN go get ./...
RUN go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@master

# copy rest of repo for build
COPY / .


#without CGO_ENABLED=0 the target environment needs glibc. Setting to zero does link statically
RUN go generate ./...
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH  go build ./cmd/server

#
# Step 2 : build frontend
#
FROM --platform=$BUILDPLATFORM node:latest AS frontend-builder
ENV NODE_ENV production
ARG REACT_APP_USER_API_BASE_URL
ARG REACT_APP_AUTH_API_BASE_URL
ARG REACT_APP_PUBLIC_URL
ARG PUBLIC_URL

WORKDIR /host-files

COPY /frontend/its-tasty/package.json .
COPY frontend/its-tasty/package-lock.json .
#Install dependencies specified in package.json
RUN npm install

COPY /frontend/its-tasty .
#afterwards build subfolder contains the static files for the frontend
RUN npm run build

#
# Step 3 : build small image with backend and frontend
#
FROM --platform=$TARGETPLATFORM alpine:latest

COPY --from=backend-builder /host-files/server /server
COPY --from=backend-builder /host-files/migrations /migrations
COPY --from=frontend-builder /host-files/build /frontend
#setup timezone
ENV TZ=Europe/Berlin
ENV ZONEINFO=/zoneinfo.zip
ENTRYPOINT ["/server"]

