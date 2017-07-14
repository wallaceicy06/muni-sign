package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/proto"
	pb "github.com/wallaceicy06/muni-sign/proto"
)

type fakeConfig struct {
	cfg    *pb.Configuration
	getErr error
	putErr error
}

func (fc *fakeConfig) Get() (*pb.Configuration, error) {
	if fc.getErr != nil {
		return nil, fc.getErr
	}
	return fc.cfg, nil
}

func (fc *fakeConfig) Put(cfg *pb.Configuration) error {
	if fc.putErr != nil {
		return fc.putErr
	}
	fc.cfg = cfg
	return nil
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

func TestGetConfig(t *testing.T) {
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
			cfg:      &fakeConfig{cfg: nil, getErr: errors.New("fake config get error")},
			req:      httptest.NewRequest(http.MethodGet, "/", &bytes.Buffer{}),
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "ConfigNil",
			cfg:      &fakeConfig{cfg: nil},
			req:      httptest.NewRequest(http.MethodGet, "/", &bytes.Buffer{}),
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := newServer(testPort, test.cfg)
			rec := &httptest.ResponseRecorder{}

			srv.rootHandler(rec, test.req)
			if rec.Code != test.wantCode {
				t.Errorf("server response code unexpected, got %d want %d", rec.Code, test.wantCode)
			}
		})
	}
}

func TestUpdateConfig(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *fakeConfig
		formAgency string
		formStopID string
		wantCode   int
		wantCfg    *pb.Configuration
	}{
		{
			name:       "Good",
			cfg:        &fakeConfig{cfg: testConfig},
			formAgency: "sf-muni",
			formStopID: "5678",
			wantCode:   http.StatusOK,
			wantCfg: &pb.Configuration{
				Agency:  "sf-muni",
				StopIds: []string{"5678"},
			},
		},
		{
			name:       "MissingAgency",
			cfg:        &fakeConfig{cfg: testConfig},
			formAgency: "",
			formStopID: "5678",
			wantCode:   http.StatusBadRequest,
			wantCfg:    testConfig,
		},
		{
			name:       "MissingStopID",
			cfg:        &fakeConfig{cfg: testConfig},
			formAgency: "sf-muni",
			formStopID: "",
			wantCode:   http.StatusBadRequest,
			wantCfg:    testConfig,
		},
		{
			name:       "ConfigPutError",
			cfg:        &fakeConfig{cfg: testConfig, putErr: errors.New("fake config put error")},
			formAgency: "sf-muni",
			formStopID: "5678",
			wantCode:   http.StatusInternalServerError,
			wantCfg:    testConfig,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := newServer(testPort, test.cfg)
			rec := &httptest.ResponseRecorder{}

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(fmt.Sprintf("agency=%s&stopId=%s", test.formAgency, test.formStopID)))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			srv.rootHandler(rec, req)
			if rec.Code != test.wantCode {
				t.Errorf("server response code unexpected, got %d want %d", rec.Code, test.wantCode)
			}
			if !proto.Equal(test.cfg.cfg, test.wantCfg) {
				t.Errorf("configurations differ: got %v, want %v", test.cfg.cfg, test.wantCfg)
			}
		})
	}
}

func TestInvalidMethod(t *testing.T) {
	srv := newServer(testPort, &fakeConfig{})
	rec := &httptest.ResponseRecorder{}

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	srv.rootHandler(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("rec.Code = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
