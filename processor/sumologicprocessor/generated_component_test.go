// Code generated by mdatagen. DO NOT EDIT.

package sumologicprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"

	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processortest"

	"go.opentelemetry.io/collector/confmap/confmaptest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/testdata"
)

// assertNoErrorHost implements a component.Host that asserts that there were no errors.
type assertNoErrorHost struct {
	component.Host
	*testing.T
}

var _ component.Host = (*assertNoErrorHost)(nil)

// newAssertNoErrorHost returns a new instance of assertNoErrorHost.
func newAssertNoErrorHost(t *testing.T) component.Host {
	return &assertNoErrorHost{
		componenttest.NewNopHost(),
		t,
	}
}

func (aneh *assertNoErrorHost) ReportFatalError(err error) {
	assert.NoError(aneh, err)
}

func TestComponentLifecycle(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name     string
		createFn func(ctx context.Context, set processor.CreateSettings, cfg component.Config) (component.Component, error)
	}{

		{
			name: "logs",
			createFn: func(ctx context.Context, set processor.CreateSettings, cfg component.Config) (component.Component, error) {
				return factory.CreateLogsProcessor(ctx, set, cfg, consumertest.NewNop())
			},
		},

		{
			name: "metrics",
			createFn: func(ctx context.Context, set processor.CreateSettings, cfg component.Config) (component.Component, error) {
				return factory.CreateMetricsProcessor(ctx, set, cfg, consumertest.NewNop())
			},
		},

		{
			name: "traces",
			createFn: func(ctx context.Context, set processor.CreateSettings, cfg component.Config) (component.Component, error) {
				return factory.CreateTracesProcessor(ctx, set, cfg, consumertest.NewNop())
			},
		},
	}

	cm, err := confmaptest.LoadConf("metadata.yaml")
	require.NoError(t, err)
	cfg := factory.CreateDefaultConfig()
	sub, err := cm.Sub("tests::config")
	require.NoError(t, err)
	require.NoError(t, component.UnmarshalConfig(sub, cfg))

	for _, test := range tests {
		t.Run(test.name+"-shutdown", func(t *testing.T) {
			c, err := test.createFn(context.Background(), processortest.NewNopCreateSettings(), cfg)
			require.NoError(t, err)
			err = c.Shutdown(context.Background())
			require.NoError(t, err)
		})

		t.Run(test.name+"-lifecycle", func(t *testing.T) {

			c, err := test.createFn(context.Background(), processortest.NewNopCreateSettings(), cfg)
			require.NoError(t, err)
			host := newAssertNoErrorHost(t)
			err = c.Start(context.Background(), host)
			require.NoError(t, err)
			assert.NotPanics(t, func() {
				switch e := c.(type) {
				case processor.Logs:
					logs := testdata.GenerateLogsManyLogRecordsSameResource(2)
					if !e.Capabilities().MutatesData {
						logs.MarkReadOnly()
					}
					err = e.ConsumeLogs(context.Background(), logs)
				case processor.Metrics:
					metrics := testdata.GenerateMetricsTwoMetrics()
					if !e.Capabilities().MutatesData {
						metrics.MarkReadOnly()
					}
					err = e.ConsumeMetrics(context.Background(), metrics)
				case processor.Traces:
					traces := testdata.GenerateTracesTwoSpansSameResource()
					if !e.Capabilities().MutatesData {
						traces.MarkReadOnly()
					}
					err = e.ConsumeTraces(context.Background(), traces)
				}
			})
			assert.NoError(t, err)
			err = c.Shutdown(context.Background())
			require.NoError(t, err)
		})
	}
}
