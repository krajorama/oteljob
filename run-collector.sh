cd /home/krajo/go/github.com/krajorama/oteljob
docker run --rm --name oteljob-collector --network host \
  -v "$(pwd)/otelcol-config.yaml:/etc/otelcol-contrib/config.yaml:ro" \
  otel/opentelemetry-collector-contrib:0.158.0
