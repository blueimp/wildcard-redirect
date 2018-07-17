FROM golang:alpine as build
WORKDIR /go/src/github.com/blueimp/wildcard-redirect
COPY . .
# Install wildcard-redirect as statically compiled binary:
# ldflags explanation (see `go tool link`):
#   -s  disable symbol table
#   -w  disable DWARF generation
RUN CGO_ENABLED=0 go install -ldflags='-s -w'

FROM scratch
COPY --from=build /go/bin/wildcard-redirect /bin/
USER 65534
ENTRYPOINT ["wildcard-redirect"]
