# Builds Prometheus from the krajo/debug-otlp-rw-endpoint branch (debug logging
# for the remote-write v1 and OTLP write receivers). Plain `go build`, no promu
# and no UI asset build: without the builtinassets tag the web UI falls back to
# serving static files from disk, so it 404s harmlessly, but every API endpoint
# -- including the write receivers this branch is about -- works fine.
FROM golang:1.26-alpine AS build
RUN apk add --no-cache git
WORKDIR /src
RUN git clone --depth 1 --branch krajo/debug-otlp-rw-endpoint https://github.com/prometheus/prometheus.git .
RUN CGO_ENABLED=0 go build -tags netgo -o /out/prometheus ./cmd/prometheus

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/prometheus /bin/prometheus
WORKDIR /prometheus
EXPOSE 9090
ENTRYPOINT ["/bin/prometheus"]
CMD ["--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.path=/prometheus"]
