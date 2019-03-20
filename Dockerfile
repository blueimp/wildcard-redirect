FROM golang:alpine as build
WORKDIR /opt
COPY . .
# Disable CGO to build a statically compiled binary.
# ldflags explanation (see `go tool link`):
#   -s  disable symbol table
#   -w  disable DWARF generation
RUN CGO_ENABLED=0 go build -ldflags='-s -w' -o /bin/wildcard-redirect

FROM scratch
COPY --from=build /bin/wildcard-redirect /bin/
USER 65534
ENTRYPOINT ["wildcard-redirect"]
