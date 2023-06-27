FROM cgr.dev/chainguard/go:1.20.5 AS builder
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
