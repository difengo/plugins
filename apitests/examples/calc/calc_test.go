package calc

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gavv/httpexpect"
	goahttp "goa.design/goa/http"
	calcsvc "goa.design/plugins/apitests/examples/calc/gen/calc"
	calcsvcsvr "goa.design/plugins/apitests/examples/calc/gen/http/calc/server"
)

func getServerHandle() http.Handler {

	// create an instance of the server
	logger := log.New(os.Stderr, "[calc] ", log.Ltime)
	svc := NewCalc(logger)
	endpoints := calcsvc.NewEndpoints(svc)
	dec := goahttp.RequestDecoder
	enc := goahttp.ResponseEncoder
	m := goahttp.NewMuxer()
	server := calcsvcsvr.New(endpoints, m, dec, enc, nil)
	calcsvcsvr.Mount(m, server)

	// returns the handler
	return m
}

func TestCalcAdd(t *testing.T) {

	// run server using httptest
	server := httptest.NewServer(getServerHandle())
	defer server.Close()

	// create httpexpect instance
	e := httpexpect.New(t, server.URL)

	// TODO: test endpoint
	/*
		r := e.GET("/add/{a}/{b}").
		WithPath("a", 1).
		WithPath("b", 2).
		Expect().
		Status(http.StatusOK).Body()

		r.Equal("3\n")
	*/
}
