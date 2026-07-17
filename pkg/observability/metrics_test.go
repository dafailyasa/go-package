package observability

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type MeterProviderSuite struct {
	suite.Suite
}

func TestMeterProviderSuite(t *testing.T) {
	suite.Run(t, new(MeterProviderSuite))
}

func (s *MeterProviderSuite) TearDownTest() {
	// Reset global provider to avoid leaking state between tests.
	otel.SetMeterProvider(sdkmetric.NewMeterProvider())
}

func (s *MeterProviderSuite) TestInitMeterProvider() {
	tests := []struct {
		name        string
		mode        string
		expectError bool
	}{
		{
			name: "Positive - OTLP HTTP",
			mode: OTLP_HTTP_MODE,
		},
		{
			name: "Positive - OTLP GRPC",
			mode: OTLP_GRPC_MODE,
		},
		{
			name:        "Negative - Invalid Mode",
			mode:        "invalid",
			expectError: true,
		},
		{
			name:        "Negative - Empty Mode",
			mode:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mp, err := InitMeterProvider(
				"user-service",
				"1.0.0",
				"test",
				tt.mode,
			)

			if tt.expectError {
				s.Nil(mp)
				s.Error(err)
				s.Equal("Invalid Observability Mode", err.Error())
				return
			}

			s.NoError(err)
			s.NotNil(mp)

			// Ensure the global provider has been set.
			s.IsType(mp, otel.GetMeterProvider())
		})
	}
}
