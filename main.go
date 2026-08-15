package main

import (
	"bufio"
	"context"
	"flag"
	"log"

	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/otlptranslator"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
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

func initMeterProvider() *sdkmetric.MeterProvider {
	var opts []prometheus.Option
	if *noTranslation {
		opts = append(opts, prometheus.WithTranslationStrategy(otlptranslator.NoTranslation))
	}
	exporter, err := prometheus.New(opts...)
	if err != nil {
		log.Fatalf("new prometheus exporter failed: %v", err)
		return nil
	}


	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(initResource()),
	)

	otel.SetMeterProvider(mp)
	return mp
}

func main() {
	flag.Parse()

	mp := initMeterProvider()
	defer func() {
		if err := mp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
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
	receivedSamples.Add(context.Background(), 1.0)


	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		receivedSamples.Add(context.Background(), 1.0)
		if scanner.Err() != nil {
			break
		}
	}
}