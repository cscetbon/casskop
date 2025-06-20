FROM golang:1.23.10 as build

ARG COMPILED_DATE
ARG VERSION

RUN useradd -u 1000 casskop
RUN mkdir -p /tmp && chown casskop /tmp

ADD . /casskop

WORKDIR /casskop

RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
    go build -ldflags "-X main.compileDate=${COMPILED_DATE}" \
             -ldflags "-X main.version=${VERSION}" \
             -a -o casskop main.go

FROM gcr.io/distroless/base

WORKDIR /
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /tmp /tmp
COPY --from=build /casskop/casskop /usr/local/bin/casskop
USER casskop

LABEL org.opencontainers.image.description="Cassandra Operator"
LABEL org.opencontainers.image.url="https://github.com/cscetbon/casskop"

ENTRYPOINT ["/usr/local/bin/casskop"]
