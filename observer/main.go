package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/otlptranslator"
	"go.opentelemetry.io/otel/attribute"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
)

var noTranslation = flag.Bool("no-translation", false, "disable OTLP-to-Prometheus metric name translation (otlptranslator.NoTranslation)")

// observedApp describes one of the applications this observer reports on.
type observedApp struct {
	serviceName string
	serviceID   string
	job         string
	instance    string
	sampleValue float64
}

func newMeterProvider(reg prometheus.Registerer, app observedApp) *sdkmetric.MeterProvider {
	resource, err := sdkresource.New(
		context.Background(),
		sdkresource.WithAttributes(
			attribute.String("service.name", app.serviceName),
			attribute.String("service.instance.id", app.serviceID),
			attribute.String("job", app.job),
			attribute.String("instance", app.instance),
		),
	)
	if err != nil {
		log.Fatalf("new resource for %s failed: %v", app.job, err)
	}

	opts := []otelprometheus.Option{
		otelprometheus.WithRegisterer(reg),
		otelprometheus.WithResourceAsConstantLabels(attribute.NewAllowKeysFilter("job", "instance")),
	}
	if *noTranslation {
		opts = append(opts, otelprometheus.WithTranslationStrategy(otlptranslator.NoTranslation))
	}
	exporter, err := otelprometheus.New(opts...)
	if err != nil {
		log.Fatalf("new prometheus exporter for %s failed: %v", app.job, err)
	}

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(resource),
	)
}

func main() {
	flag.Parse()

	apps := []observedApp{
		{serviceName: "my_service1", serviceID: "my_id1", job: "my_job1", instance: "my_instance1", sampleValue: 1},
		{serviceName: "my_service2", serviceID: "my_id2", job: "my_job2", instance: "my_instance2", sampleValue: 2},
	}

	registry := prometheus.NewRegistry()

	var providers []*sdkmetric.MeterProvider
	for _, app := range apps {
		mp := newMeterProvider(registry, app)
		providers = append(providers, mp)

		meter := mp.Meter("foo")
		receivedSamples, err := meter.Float64Counter("received_samples")
		if err != nil {
			log.Fatalf("new counter for %s failed: %v", app.job, err)
		}
		receivedSamples.Add(context.Background(), app.sampleValue, metric.WithAttributes(attribute.String("a.a", "b")))
	}

	defer func() {
		for _, mp := range providers {
			if err := mp.Shutdown(context.Background()); err != nil {
				log.Printf("Error shutting down meter provider: %v", err)
			}
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":9464", nil))
}
