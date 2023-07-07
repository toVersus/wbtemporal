FROM cgr.dev/chainguard/go@sha256:8ed3fdc8f6375a3fd84b4b8b696a2366c3a639931aab492d6f92ca917e726ad6 AS builder
WORKDIR /go/src/github.com/toVersus/wbtemporal
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/wbtemporal .

FROM cgr.dev/chainguard/wolfi-base:latest
WORKDIR /

ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    wbctl
USER wbctl

COPY --from=builder /bin/wbtemporal /usr/local/bin/
ENTRYPOINT ["wbtemporal"]
