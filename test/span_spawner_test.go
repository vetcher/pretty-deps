package test

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"syreclabs.com/go/faker"
)

const (
	opentracingHost = "localhost"
	opentracingPort = 9411
)

const (
	spanAmount    = 50
	serviceAmount = 40
	maxDeep       = 5
	minDeep       = 1
	nextRatio     = 0.15
)

type sAndT struct {
	service string
	tracer  opentracing.Tracer
}

var (
	services  []sAndT
	collector zipkin.Collector
)

func TestMain(m *testing.M) {
	connectToCollector()
	services = make([]sAndT, serviceAmount)
	for i := range services {
		services[i] = connectToTracer(strings.Replace(faker.App().Name(), " ", "", -1))
	}
	m.Run()
}

func connectToCollector() {
	connString := fmt.Sprintf("http://%s:%d/api/v1/spans", opentracingHost, opentracingPort)
	c, err := zipkin.NewHTTPCollector(connString)
	if err != nil {
		panic(err)
	}
	collector = c
}

func connectToTracer(name string) sAndT {
	recorder := zipkin.NewRecorder(collector, false, "0.0.0.0:9000", name)
	tracer, err := zipkin.NewTracer(recorder, zipkin.ClientServerSameSpan(true), zipkin.TraceID128Bit(true))
	if err != nil {
		panic(err)
	}
	return sAndT{
		tracer:  tracer,
		service: name,
	}
}

func TestSpawnSpans(t *testing.T) {
	var wg sync.WaitGroup
	fmt.Println("generating spans")
	for i := 0; i < spanAmount; i++ {
		wg.Add(1)
		spawnSpan(&wg, nil, 0)
	}
	wg.Wait()
	fmt.Println("done")
	collector.Close()
}

func spawnSpan(wg *sync.WaitGroup, parent opentracing.Span, deep int) {
	s := services[rand.Intn(serviceAmount)]
	var opts []opentracing.StartSpanOption
	if parent != nil {
		opts = append(opts, opentracing.ChildOf(parent.Context()))
	}
	span := s.tracer.StartSpan("", opts...)
	ext.SpanKindRPCClient.Set(span)
	go func() {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(2000)))
		ss := services[rand.Intn(serviceAmount)]
		serverSpan := ss.tracer.StartSpan("", ext.RPCServerOption(span.Context()))
		ext.SpanKindRPCServer.Set(serverSpan)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(2000)))
		serverSpan.Finish()
		span.Finish()
		wg.Done()
	}()
	if deep < maxDeep {
		if deep < minDeep || rand.Float64() < nextRatio {
			wg.Add(1)
			spawnSpan(wg, span, deep+1)
		}
	}
}
