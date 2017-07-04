package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/wallaceicy06/muni-sign/proto"
)

type fakeConfig struct {
	cfg *pb.Configuration
	err error
}

func (fc *fakeConfig) Get() (*pb.Configuration, error) {
	return fc.cfg, fc.err
}

func (fc *fakeConfig) Put(*pb.Configuration) error {
	return fc.err
}

const testPort = 25565

var testConfig = &pb.Configuration{
	Agency:  "sf-muni",
	StopIds: []string{"1234", "5678"},
}

func TestServing(t *testing.T) {
	srv := newServer(testPort, &fakeConfig{cfg: testConfig}).serve()
	defer srv.Shutdown(context.Background())

	res, err := http.Get(fmt.Sprintf("http://localhost:%d", testPort))
	if err != nil {
		t.Fatalf("problem starting server: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected OK response from server, got %d want %d", res.StatusCode, http.StatusOK)
	}
}

func TestRootHandler(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *fakeConfig
		req      *http.Request
		wantCode int
	}{
		{
			name:     "Good",
			cfg:      &fakeConfig{cfg: testConfig},
			req:      httptest.NewRequest(http.MethodGet, "/", &bytes.Buffer{}),
			wantCode: http.StatusOK,
		},
		{
			name:     "ConfigError",
			cfg:      &fakeConfig{cfg: nil, err: errors.New("fake config error")},
			req:      httptest.NewRequest(http.MethodGet, "/", &bytes.Buffer{}),
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "ConfigNil",
			cfg:      &fakeConfig{cfg: nil, err: nil},
			req:      httptest.NewRequest(http.MethodGet, "/", &bytes.Buffer{}),
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := newServer(testPort, test.cfg)
			rec := &httptest.ResponseRecorder{}

			req := httptest.NewRequest(http.MethodGet, "/", &bytes.Buffer{})

			srv.rootHandler(rec, req)
			if rec.Code != test.wantCode {
				t.Errorf("server response code unexpected, got %d want %d", rec.Code, test.wantCode)
			}
		})
	}
}
