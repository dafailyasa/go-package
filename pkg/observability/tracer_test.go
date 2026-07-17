package observability

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ObservabilitySuite struct {
	suite.Suite
}

func TestObservabilitySuite(t *testing.T) {
	suite.Run(t, new(ObservabilitySuite))
}

func (s *ObservabilitySuite) TearDownTest() {
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())
}

func (s *ObservabilitySuite) TestInitTracerProvider_Stdout() {
	tp, err := InitTracerProvider(
		"test-app",
		"1.0.0",
		"test",
		"",
	)

	s.NoError(err)
	s.NotNil(tp)

	s.IsType(tp, otel.GetTracerProvider())
	s.NotNil(otel.GetTextMapPropagator())
}

func (s *ObservabilitySuite) TestInitTracerProvider_UnknownMode() {
	tp, err := InitTracerProvider(
		"test-app",
		"1.0.0",
		"test",
		"unknown-mode",
	)

	s.NoError(err)
	s.NotNil(tp)

	s.IsType(tp, otel.GetTracerProvider())
}

func (s *ObservabilitySuite) TestInitTracerProvider_OTLPHTTP() {
	tp, err := InitTracerProvider(
		"test-app",
		"1.0.0",
		"test",
		OTLP_HTTP_MODE,
	)

	s.NoError(err)
	s.NotNil(tp)
}

func (s *ObservabilitySuite) TestInitTracerProvider_OTLPGRPC() {
	tp, err := InitTracerProvider(
		"test-app",
		"1.0.0",
		"test",
		OTLP_GRPC_MODE,
	)

	s.NoError(err)
	s.NotNil(tp)
}
