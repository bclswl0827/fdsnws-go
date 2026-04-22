package provider

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	fdsnws "github.com/bclswl0827/fdsnws-go"
	"github.com/bclswl0827/mseedio"
)

func TestDataSelectProducesReadableMiniSEED(t *testing.T) {
	server, err := fdsnws.NewFDSNWS(New(), fdsnws.WithModules(fdsnws.ModuleDataSelect))
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	end := start.Add(5 * time.Second)
	request := httptest.NewRequest(http.MethodGet,
		"/fdsnws/dataselect/1/query?net=XX&sta=DEMO&loc=00&cha=BHZ&start=2024-01-02T03:04:05&end=2024-01-02T03:04:10", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var decoded mseedio.MiniSeedData
	if err := decoded.ReadFromReader(bytes.NewReader(recorder.Body.Bytes())); err != nil {
		t.Fatalf("decode generated miniSEED: %v", err)
	}
	if decoded.Records != 5 || decoded.Samples != 5*sampleRate {
		t.Fatalf("decoded = %#v", decoded)
	}
	if !decoded.StartTime.Equal(start) {
		t.Fatalf("start time = %s, want %s", decoded.StartTime, start)
	}
	last := decoded.Series[len(decoded.Series)-1].FixedSection
	decodedEnd := last.StartTime.Add(time.Duration(last.SamplesNumber) * time.Second / sampleRate)
	if !decodedEnd.Equal(end) {
		t.Fatalf("end time = %s, want %s", decodedEnd, end)
	}
}

func TestDataSelectRejectsLongRangeAsClientError(t *testing.T) {
	server, err := fdsnws.NewFDSNWS(New(), fdsnws.WithModules(fdsnws.ModuleDataSelect))
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet,
		"/fdsnws/dataselect/1/query?net=XX&sta=DEMO&loc=00&cha=BHZ&start=2024-01-02T03:04:05&end=2024-01-02T05:04:05", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "time range too long") {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}
