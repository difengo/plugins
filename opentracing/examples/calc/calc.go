package calc

import (
	"context"
	"log"

	opentracing "github.com/opentracing/opentracing-go"
	calcsvc "goa.design/plugins/opentracing/examples/calc/gen/calc"
)

// calc service example implementation.
// The example methods log the requests and return zero values.
type calcSvc struct {
	logger *log.Logger
}

// NewCalc returns the calc service implementation.
func NewCalc(logger *log.Logger) calcsvc.Service {
	return &calcSvc{logger}
}

// Add implements add.
func (s *calcSvc) Add(ctx context.Context, p *calcsvc.AddPayload) (res int, err error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "calc.Add")
	defer span.Finish()
	s.logger.Print("calc.add")
	return
}
