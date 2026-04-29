package main

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// initTracer inicializa e configura o rastreamento OpenTelemetry
func initTracer() *sdktrace.TracerProvider {
	ctx := context.Background()

	// Lê a env var injetada pelo GitOps
	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		log.Println("Aviso: OTEL_EXPORTER_OTLP_ENDPOINT não está definido. Traces não serão exportados.")
		// Opcional: retornar um NoopTracer se não houver endpoint
	}

	// 1. Configura o Exporter (Envia os Traces via HTTP/Protobuf)
	// Como o endpoint já vem com o "http://" no K8s configmap, precisamos remover o scheme 
	// ou usar otlptracehttp.WithEndpointURL
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithInsecure(), // Comunicação interna K8s não precisa de TLS
		otlptracehttp.WithEndpointURL(otelEndpoint),
	)
	if err != nil {
		log.Fatalf("Falha ao criar o exporter OTLP: %v", err)
	}

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "auth-service-unknown"
	}

	// 2. Cria o Resource (Metadados do Serviço)
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		log.Fatalf("Falha ao criar o OTel resource: %v", err)
	}

	// 3. Cria o TracerProvider em Batch
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Registra o TracerProvider como global
	otel.SetTracerProvider(tp)

	// Configura o formato de propagação de contexto (W3C Trace Context)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	log.Println("OpenTelemetry TracerProvider inicializado com sucesso!")
	return tp
}
