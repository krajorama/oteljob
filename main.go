package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"maps"

	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/otlptranslator"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
)

var noTranslation = flag.Bool("no-translation", false, "disable OTLP-to-Prometheus metric name translation (otlptranslator.NoTranslation)")

var resource *sdkresource.Resource
var initResourcesOnce sync.Once

func initResource() *sdkresource.Resource {
	initResourcesOnce.Do(func() {
		extraResources, _ := sdkresource.New(
			context.Background(),
			// sdkresource.WithOS(),
			// sdkresource.WithProcess(),
			// sdkresource.WithContainer(),
			// sdkresource.WithHost(),
			sdkresource.WithAttributes(attribute.String("service.name", "my_service"), attribute.String("service.instance.id", "my_id")),
		)
		resource, _ = sdkresource.Merge(
			sdkresource.Default(),
			extraResources,
		)
	})
	return resource
}

func initMeterProvider() (*sdkmetric.MeterProvider, sdkmetric.Reader) {
	var opts []prometheus.Option
	if *noTranslation {
		opts = append(opts, prometheus.WithTranslationStrategy(otlptranslator.NoTranslation))
	}
	exporter, err := prometheus.New(opts...)
	if err != nil {
		log.Fatalf("new prometheus exporter failed: %v", err)
		return nil, nil
	}


	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(initResource()),
	)

	otel.SetMeterProvider(mp)
	return mp, exporter
}

// logOTelModel dumps the raw OTel data model as the SDK sees it: resource
// attributes, and for each meter (scope) its instruments' data points with
// their attributes, before any Prometheus-specific translation happens.
func logOTelModel(ctx context.Context, reader sdkmetric.Reader) {
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		log.Printf("collecting OTel model failed: %v", err)
		return
	}

	log.Printf("resource attributes: %s", rm.Resource.Encoded(attribute.DefaultEncoder()))

	for _, sm := range rm.ScopeMetrics {
		log.Printf("meter %q (version %q)", sm.Scope.Name, sm.Scope.Version)
		for _, m := range sm.Metrics {
			switch data := m.Data.(type) {
			case metricdata.Gauge[float64]:
				for _, dp := range data.DataPoints {
					log.Printf("  %s{%s} = %v", m.Name, dp.Attributes.Encoded(attribute.DefaultEncoder()), dp.Value)
				}
			case metricdata.Sum[float64]:
				for _, dp := range data.DataPoints {
					log.Printf("  %s{%s} = %v", m.Name, dp.Attributes.Encoded(attribute.DefaultEncoder()), dp.Value)
				}
			case metricdata.Histogram[float64]:
				for _, dp := range data.DataPoints {
					log.Printf("  %s{%s} count=%d sum=%v", m.Name, dp.Attributes.Encoded(attribute.DefaultEncoder()), dp.Count, dp.Sum)
				}
			default:
				log.Printf("  %s: unhandled aggregation type %T", m.Name, m.Data)
			}
		}
	}
}

// filteredMetricsPrefixes lists metric family prefixes to omit from the
// "/metrics response" debug log, to cut down noise from Go runtime and
// promhttp's own instrumentation.
var filteredMetricsPrefixes = []string{"go_", "process_", "promhttp_"}

// metricNameFromLine extracts the metric name from a line of Prometheus text
// exposition format, or "" if the line has no associated metric name.
func metricNameFromLine(line string) string {
	switch {
	case strings.HasPrefix(line, "# HELP "), strings.HasPrefix(line, "# TYPE "):
		fields := strings.SplitN(line, " ", 4)
		if len(fields) >= 3 {
			return fields[2]
		}
		return ""
	case line == "", strings.HasPrefix(line, "#"):
		return ""
	default:
		return line[:strings.IndexAny(line+"{", "{ ")]
	}
}

func filterMetricsForLog(body string) string {
	lines := strings.Split(body, "\n")
	kept := lines[:0]
	for _, line := range lines {
		name := metricNameFromLine(line)
		skip := false
		for _, prefix := range filteredMetricsPrefixes {
			if strings.HasPrefix(name, prefix) {
				skip = true
				break
			}
		}
		if !skip {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

func main() {
	flag.Parse()

	mp, reader := initMeterProvider()
	defer func() {
		if err := mp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
		}
	}()

	metricsHandler := promhttp.Handler()
	http.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("/metrics called by %s, Accept: %s", r.RemoteAddr, r.Header.Get("Accept"))
		logOTelModel(r.Context(), reader)
		rec := httptest.NewRecorder()
		metricsHandler.ServeHTTP(rec, r)
		if rec.Code == http.StatusOK {
			log.Printf("/metrics response:\n%s\nend response", filterMetricsForLog(rec.Body.String()))
		}
		maps.Copy(w.Header(), rec.Header())
		w.WriteHeader(rec.Code)
		w.Write(rec.Body.Bytes())
	}))
	go func() {
		if err := http.ListenAndServe(":9464", nil); err != nil {
			log.Fatalf("metrics server failed: %v", err)
		}
	}()

	meter := mp.Meter("foo")

	receivedSamples, err := meter.Float64Counter("received_samples")
	if err != nil {
		log.Fatalf("new histogram failed: %v", err)
		return
	}
	sampleAttrs := metric.WithAttributes(attribute.String("a.a", "b"))
	receivedSamples.Add(context.Background(), 1.0, sampleAttrs)


	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		sampleAttrs := metric.WithAttributes(attribute.String("a.a", "b"))
	receivedSamples.Add(context.Background(), 1.0, sampleAttrs)
		if scanner.Err() != nil {
			break
		}
	}
}