package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/wallaceicy06/muni-sign/proto"
	grpcContext "golang.org/x/net/context"
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

type fakeNbClient struct {
	agenciesRes *pb.ListAgenciesResponse
	agenciesErr error
}

func (fnb *fakeNbClient) ListAgencies(ctx grpcContext.Context, req *pb.ListAgenciesRequest, _ ...grpc.CallOption) (*pb.ListAgenciesResponse, error) {
	if fnb.agenciesErr != nil {
		return nil, fnb.agenciesErr
	}
	return fnb.agenciesRes, nil
}

func (fnb *fakeNbClient) ListPredictions(ctx grpcContext.Context, req *pb.ListPredictionsRequest, _ ...grpc.CallOption) (*pb.ListPredictionsResponse, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "Fake ListPredictions is unimplemented.")
}

const testPort = 25565

var testConfig = &pb.Configuration{
	Agency:  "sf-muni",
	StopIds: []string{"1234", "5678"},
}

var goodFakeNb = &fakeNbClient{
	agenciesRes: &pb.ListAgenciesResponse{
		Agencies: []*pb.Agency{{Name: "San Francisco MTA", Tag: "sf-muni"}},
	}}

func TestServing(t *testing.T) {
	srv := newServer(testPort, goodFakeNb, &fakeConfig{cfg: testConfig}).serve()
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
			srv := newServer(testPort, goodFakeNb, test.cfg)
			rec := httptest.NewRecorder()

			srv.rootHandler(rec, test.req)
			res := rec.Result()

			if res.StatusCode != test.wantCode {
				t.Errorf("server response code unexpected, got %d want %d", res.StatusCode, test.wantCode)
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
			name:       "OneStop",
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
			name:       "MultipleStops",
			cfg:        &fakeConfig{cfg: testConfig},
			formAgency: "sf-muni",
			formStopID: "1234 5678 9012",
			wantCode:   http.StatusOK,
			wantCfg: &pb.Configuration{
				Agency:  "sf-muni",
				StopIds: []string{"1234", "5678", "9012"},
			},
		},
		{
			name:       "MultipleStopsExtraSpaces",
			cfg:        &fakeConfig{cfg: testConfig},
			formAgency: "sf-muni",
			formStopID: "      1234  5678        9012  ",
			wantCode:   http.StatusOK,
			wantCfg: &pb.Configuration{
				Agency:  "sf-muni",
				StopIds: []string{"1234", "5678", "9012"},
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
			name:       "EmptyStopIds",
			cfg:        &fakeConfig{cfg: testConfig},
			formAgency: "sf-muni",
			formStopID: "",
			wantCode:   http.StatusOK,
			wantCfg:    &pb.Configuration{Agency: "sf-muni"},
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
			srv := newServer(testPort, goodFakeNb, test.cfg)
			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(fmt.Sprintf("agency=%s&stopIds=%s", test.formAgency, test.formStopID)))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			srv.rootHandler(rec, req)
			res := rec.Result()

			if res.StatusCode != test.wantCode {
				t.Errorf("server response code unexpected, got %d want %d", res.StatusCode, test.wantCode)
			}
			if !proto.Equal(test.cfg.cfg, test.wantCfg) {
				t.Errorf("configurations differ: got %v, want %v", test.cfg.cfg, test.wantCfg)
			}
		})
	}
}

func TestRootInvalidMethod(t *testing.T) {
	srv := newServer(testPort, goodFakeNb, &fakeConfig{})
	rec := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	srv.rootHandler(rec, req)
	res := rec.Result()

	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("StatusCode = %d, want %d", res.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestApiConfigGet(t *testing.T) {
	tests := []struct {
		name    string
		fakeCfg *fakeConfig
		wantErr bool
	}{
		{
			name:    "Good",
			fakeCfg: &fakeConfig{cfg: testConfig},
			wantErr: false,
		},
		{
			name:    "Error",
			fakeCfg: &fakeConfig{getErr: errors.New("fake config get error")},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := newServer(testPort, goodFakeNb, test.fakeCfg)
			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
			srv.apiConfigHandler(rec, req)
			res := rec.Result()

			if test.wantErr {
				if res.StatusCode == http.StatusOK {
					t.Error("get API config got OK, want error")
				}
				return
			}

			if res.StatusCode != http.StatusOK {
				t.Errorf("get API config got code %d want %d", res.StatusCode, http.StatusOK)
			}

			got := &pb.Configuration{}
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("error reading API config response: %v", err)
			}
			if err := json.Unmarshal(body, got); err != nil {
				t.Fatalf("error unmarshaling JSON response: %v", err)
			}

			if !proto.Equal(got, test.fakeCfg.cfg) {
				t.Errorf("configuration does not match: got %v want %v", got, test.fakeCfg.cfg)
			}
		})
	}
}

func TestApiConfigInvalidMethod(t *testing.T) {
	srv := newServer(testPort, goodFakeNb, &fakeConfig{})
	rec := &httptest.ResponseRecorder{}

	req := httptest.NewRequest(http.MethodPost, "/api/config", nil)
	srv.apiConfigHandler(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("rec.Code = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestAgencyListCache(t *testing.T) {
	lastRefresh := time.Now()

	cachedAgencies := []*pb.Agency{{
		Name: "Los Angeles Metro",
		Tag:  "la-metro",
	}}

	tests := []struct {
		name    string
		timeNow time.Time
		fakeNb  *fakeNbClient
		want    []*pb.Agency
	}{
		{
			name:    "FetchFromServer",
			timeNow: lastRefresh.Add(cacheTimeout + time.Second),
			fakeNb:  goodFakeNb,
			want: []*pb.Agency{{
				Name: "San Francisco MTA",
				Tag:  "sf-muni",
			}},
		},
		{
			name:    "FetchFromServerError",
			timeNow: lastRefresh.Add(cacheTimeout + time.Second),
			fakeNb: &fakeNbClient{
				agenciesErr: errors.New("fake list agencies error"),
			},
			want: []*pb.Agency{{
				Name: "San Francisco MTA",
				Tag:  "sf-muni",
			}},
		},
		{
			name:    "FetchFromCache",
			timeNow: lastRefresh.Add(cacheTimeout - time.Second),
			fakeNb:  goodFakeNb,
			want:    cachedAgencies,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := newServer(testPort, test.fakeNb, &fakeConfig{})
			srv.agencyCache = &agencyCache{
				lastRefresh: lastRefresh,
				agencies:    cachedAgencies,
			}
			timeNow = func() time.Time { return test.timeNow }

			got := srv.getAgencies()
			if len(got) != len(test.want) {
				t.Fatalf("got %d agencies, want %d", len(got), len(test.want))
			}

			mismatch := false
			for i, gotAgency := range got {
				mismatch = mismatch && !proto.Equal(gotAgency, test.want[i])
			}

			if mismatch {
				t.Errorf("agencies differ: got %v want %v", got, test.want)
			}
		})
	}
}
