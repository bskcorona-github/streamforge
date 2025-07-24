package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitTracer initializes the OpenTelemetry tracer
func InitTracer(endpoint string) (*sdktrace.TracerProvider, error) {
	if endpoint == "" {
		endpoint = "localhost:4317"
	}

	// Create gRPC connection
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("streamforge-collector"),
			semconv.ServiceVersion("0.1.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace exporter
	traceExporter, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create metric exporter
	metricExporter, err := otlpmetricgrpc.New(context.Background(), otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	// Create metric provider
	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(res),
	)

	// Set global providers
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	return tp, nil
}

// CreateSpan creates a new span with the given name and attributes
func CreateSpan(ctx context.Context, name string, attributes ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := otel.Tracer("streamforge-collector")
	return tracer.Start(ctx, name, attributes...)
}

// RecordMetric records a metric with the given name and value
func RecordMetric(ctx context.Context, name string, value float64, attributes ...string) {
	meter := otel.Meter("streamforge-collector")
	counter, err := meter.Float64Counter(name)
	if err != nil {
		// Log error but don't fail
		return
	}
	counter.Add(ctx, value)
}

// RecordHistogram records a histogram metric
func RecordHistogram(ctx context.Context, name string, value float64, attributes ...string) {
	meter := otel.Meter("streamforge-collector")
	histogram, err := meter.Float64Histogram(name)
	if err != nil {
		// Log error but don't fail
		return
	}
	histogram.Record(ctx, value)
}

// RecordGauge records a gauge metric
func RecordGauge(ctx context.Context, name string, value float64, attributes ...string) {
	meter := otel.Meter("streamforge-collector")
	gauge, err := meter.Float64ObservableGauge(name)
	if err != nil {
		// Log error but don't fail
		return
	}
	
	// Note: Gauge recording requires a callback, this is a simplified version
	// In a real implementation, you would use the callback mechanism
	_ = gauge
} 