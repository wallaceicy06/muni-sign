package config

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/afero"

	pb "github.com/wallaceicy06/muni-sign/proto"
)

func TestGet(t *testing.T) {
	goodFilePath := "/path/to/file"
	badFilePath := "/this/file/is/bad"

	goodCfg := `agency: "sf-muni"
			   stop_ids: "1234"`

	tests := []struct {
		name     string
		filePath string
		fileData string
		wantCfg  *pb.Configuration
		wantErr  bool
	}{
		{
			name:     "Simple",
			filePath: goodFilePath,
			fileData: goodCfg,
			wantCfg: &pb.Configuration{
				Agency:  "sf-muni",
				StopIds: []string{"1234"},
			},
		},
		{
			name:     "MultipleStops",
			filePath: goodFilePath,
			fileData: `agency: "sf-muni"
					   stop_ids: "1234",
					   stop_ids: "5678"`,
			wantCfg: &pb.Configuration{
				Agency:  "sf-muni",
				StopIds: []string{"1234", "5678"},
			},
		},
		{
			name:     "InvalidPath",
			filePath: badFilePath,
			fileData: goodCfg,
			wantErr:  true,
		},

		{
			name:     "InvalidConfig",
			filePath: goodFilePath,
			fileData: `agency: 1234`,
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs = afero.NewMemMapFs()
			afero.WriteFile(fs, goodFilePath, []byte(test.fileData), 0644)

			sc := NewSignConfig(test.filePath)
			got, err := sc.Get()

			if test.wantErr {
				if err == nil {
					t.Errorf("sc.Get() = _, <nil> want _, <non-nil>")
				}
				return
			}

			if err != nil {
				t.Errorf("sc.Get() = _, %v want _, <nil>", err)
			}
			if !proto.Equal(got, test.wantCfg) {
				t.Errorf("sc.Get() = %v, _ want %v, _", got, test.wantCfg)
			}
		})
	}
}

func TestPut(t *testing.T) {
	goodFilePath := "/path/to/file"

	goodCfg := &pb.Configuration{
		Agency:  "sf-muni",
		StopIds: []string{"1234"},
	}

	tests := []struct {
		name         string
		cfg          *pb.Configuration
		filePath     string
		wantFileData string
		wantErr      bool
	}{
		{
			name:     "Simple",
			cfg:      goodCfg,
			filePath: goodFilePath,
		},
		{
			name: "MultipleStops",
			cfg: &pb.Configuration{
				Agency:  "sf-muni",
				StopIds: []string{"1234", "5678"},
			},
			filePath: goodFilePath,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs = afero.NewMemMapFs()

			sc := NewSignConfig(test.filePath)
			err := sc.Put(test.cfg)

			if test.wantErr {
				if err == nil {
					t.Errorf("sc.Put() = <nil> want <non-nil>")
				}
				return
			}

			if err != nil {
				t.Errorf("sc.Put() = %v want <nil>", err)
			}

			gotBytes, err := afero.ReadFile(fs, goodFilePath)
			if err != nil {
				t.Errorf("error reading config file: %v", err)
			}
			want := proto.MarshalTextString(test.cfg)
			got := string(gotBytes)
			if got != want {
				t.Errorf("config files differ: got %v want %v", got, want)
			}
		})
	}
}
