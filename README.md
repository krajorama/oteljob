# oteljob

A docker-compose sandbox for exercising how Prometheus scrape identity
(`job`/`instance`) survives OTel SDK → Collector → Prometheus translation,
built around the use cases in [UseCases.md](UseCases.md) and verified against
the [`krajo/job-instance-persistence`](https://github.com/krajorama/opentelemetry-collector-contrib/tree/krajo/job-instance-persistence)
collector branch and the [`krajo/debug-otlp-rw-endpoint`](https://github.com/prometheus/prometheus/tree/krajo/debug-otlp-rw-endpoint)
Prometheus branch. See [verification-default-translation.md](verification-default-translation.md)
and [verification-notranslation.md](verification-notranslation.md) (or their
`.html` renders) for the results.

## You must select a Compose profile

`docker compose up` on its own starts nothing app-specific — `oteljob` and
`observer` are mutually exclusive use cases, gated behind Compose profiles, so
you always need `--profile`:

```sh
docker compose --profile standalone up -d   # use case 1/2: single app
docker compose --profile observer up -d     # use case 3: one process observing two apps
```

`otel-collector` and `prometheus` belong to both profiles and start with
either one. Nothing stops you from starting both profiles at once (no port or
name conflicts), but it mixes two independent test scenarios' data into the
same Prometheus, which usually isn't what you want when comparing results.

Tear down with the same `--profile` flag you started with, e.g.
`docker compose --profile standalone down`.

## Layout

- `standalone/` — single OTel SDK app (`oteljob`), use cases 1 and 2 (default
  vs. `NoTranslation` translation strategy).
- `observer/` — one process reporting on two synthetic observed apps via two
  `MeterProvider`s sharing one Prometheus registry, use case 3.
- `otelcollector/` — builds the collector from the `krajo/job-instance-persistence`
  fork (`otelcol-config.compose.yaml` is its config).
- `prometheus/` — builds Prometheus from the `krajo/debug-otlp-rw-endpoint`
  fork.

## Useful environment variables

| Variable | Effect |
| :--- | :--- |
| `OTELJOB_NO_TRANSLATION=true` | `standalone`: enable the SDK's `NoTranslation` strategy |
| `OBSERVER_NO_TRANSLATION=true` | `observer`: same, for both observed apps |
| `OTELCOL_FEATURE_GATES=<gate1,gate2>` | Enable collector feature gates, e.g. `receiver.prometheusreceiver.jobInstanceOptionA,exporter.prometheusremotewrite.jobInstanceOptionA` |
| `OTELCOL_RW_TRANSLATION_STRATEGY=<strategy>` | `prometheus_remote_write` exporter's `translation_strategy`. Default `UnderscoreEscapingWithSuffixes` (matches today's implicit default); set to `NoTranslation` for use cases 2/4 |
| `OTELCOL_RW_PROTOBUF_MESSAGE=<message>` | `prometheus_remote_write` exporter's `protobuf_message`. Default `prometheus.WriteRequest` (Remote Write 1.0); must be `io.prometheus.write.v2.Request` (Remote Write 2.0) whenever `OTELCOL_RW_TRANSLATION_STRATEGY` is `NoUTF8EscapingWithSuffixes` or `NoTranslation` — the exporter rejects those strategies over RW 1.0 |

Set both the app's translation flag and the collector's feature gates in the
same `docker compose up` invocation — Compose only applies environment
variables present at the time a container is (re)created, so setting one and
recreating in a second command silently reverts the other to its default.

For use cases 2/4 (no translation end-to-end), all of the following must be
set together in the same invocation: `OTELJOB_NO_TRANSLATION=true` (or
`OBSERVER_NO_TRANSLATION=true`), `OTELCOL_RW_TRANSLATION_STRATEGY=NoTranslation`,
`OTELCOL_RW_PROTOBUF_MESSAGE=io.prometheus.write.v2.Request`, and
`OTELCOL_FEATURE_GATES=exporter.prometheusremotewritexporter.enableSendingRW2`
(RW 2.0 sending is gated behind that alpha feature gate).

## Regenerating the HTML reports

```sh
make html        # verification-*.md -> verification-*.html
make clean-html
```
