package cmd

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var (
	meter                   metric.Meter
	policyComplianceCounter metric.Int64Counter
	policyComplianceGauge   metric.Int64UpDownCounter
	sarifLevelGauge         metric.Int64ObservableGauge
	exporter                *prometheus.Exporter
	sarifLevels             = make(map[string]int64)
	sarifLevelsMutex        sync.RWMutex
)

func InitOpenTelemetry() error {
	var err error
	exporter, err = prometheus.New()
	if err != nil {
		return fmt.Errorf("failed to create Prometheus exporter: %v", err)
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(semconv.ServiceNameKey.String("intercept-cli")),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %v", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)
	otel.SetMeterProvider(meterProvider)

	meter = meterProvider.Meter("intercept-metrics")

	policyComplianceCounter, err = meter.Int64Counter("policy_compliance_total",
		metric.WithDescription("Total number of policy compliance checks"),
	)
	if err != nil {
		return fmt.Errorf("failed to create policy compliance counter: %v", err)
	}

	policyComplianceGauge, err = meter.Int64UpDownCounter("policy_compliance_state",
		metric.WithDescription("Current state of policy compliance (1 for compliant, -1 for non-compliant)"),
	)
	if err != nil {
		return fmt.Errorf("failed to create policy compliance gauge: %v", err)
	}

	sarifLevelGauge, err = meter.Int64ObservableGauge("sarif_level",
		metric.WithDescription("Current SARIF level for each policy"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			sarifLevelsMutex.RLock()
			defer sarifLevelsMutex.RUnlock()
			for policyID, level := range sarifLevels {
				o.Observe(level, metric.WithAttributes(attribute.String("policy_id", policyID)))
			}
			return nil
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create SARIF level gauge: %v", err)
	}

	return nil
}

func StartMetricsServer(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}

func RecordPolicyCompliance(ctx context.Context, policyID string, compliant bool) {
	policyComplianceCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("policy_id", policyID),
			attribute.Bool("compliant", compliant),
		),
	)

	if compliant {
		policyComplianceGauge.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("policy_id", policyID),
			),
		)
	} else {
		policyComplianceGauge.Add(ctx, -1,
			metric.WithAttributes(
				attribute.String("policy_id", policyID),
			),
		)
	}
}

func RecordSarifLevel(ctx context.Context, policyID string, sarifLevel string, sarifLevelInt int) {
	sarifLevelsMutex.Lock()
	sarifLevels[policyID] = int64(sarifLevelInt)
	sarifLevelsMutex.Unlock()
}
