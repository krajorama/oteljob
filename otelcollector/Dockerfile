# Builds a minimal otelcol-contrib from the krajo/job-instance-persistence
# fork: just the components this compose stack actually uses (prometheus
# receiver, debug/otlphttp/prometheusremotewrite exporters), with the three
# modules that branch touches (receiver/prometheusreceiver,
# exporter/prometheusremotewriteexporter, pkg/translator/prometheusremotewrite)
# replaced with the fork checkout via the OCB manifest's `replaces:` section.
FROM golang:1.26-alpine AS build
RUN apk add --no-cache git
RUN git clone --depth 1 --branch krajo/job-instance-persistence https://github.com/krajorama/opentelemetry-collector-contrib.git /contrib-fork

ENV CGO_ENABLED=0
ENV PATH="/root/go/bin:${PATH}"
RUN go install -trimpath -ldflags="-s -w" go.opentelemetry.io/collector/cmd/builder@v0.158.0

WORKDIR /src
COPY manifest.yaml .
RUN builder --config=manifest.yaml

FROM gcr.io/distroless/static-debian12
COPY --from=build /src/_build/otelcol-custom /otelcol-contrib
ENTRYPOINT ["/otelcol-contrib"]
CMD ["--config=/etc/otelcol-contrib/config.yaml"]
