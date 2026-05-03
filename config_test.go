package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIsValidRegion(t *testing.T) {
	tests := []struct {
		region string
		valid  bool
	}{
		{"us-east-1", true},
		{"eu-west-1", true},
		{"all", false},
		{"", false},
		{"invalid-region", false},
		{"US-EAST-1", false},
		{"us-east-99", false},
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			got := isValidRegion(tt.region)
			if got != tt.valid {
				t.Errorf("isValidRegion(%q) = %v, want %v", tt.region, got, tt.valid)
			}
		})
	}
}

func TestOptionsValidate(t *testing.T) {
	// overwrite VALID_REGIONS to make it easier to write tests
	VALID_REGIONS = map[string]string{
		"eu-west-1": "Europe (Ireland)",
		"us-east-1": "US East (N. Virginia)",
		"us-east-2": "US East (Ohio)",
	}

	tests := []struct {
		name              string
		config            NoclickopsConfig
		wantErr           bool
		errContains       string
		wantRegionsList   []string
		wantIgnoreTagsMap map[string][]string
	}{
		{
			name:        "invalid with statefile only regions all",
			config:      NoclickopsConfig{stateFile: "state.tfstate", regions: "all"},
			wantErr:     true,
			errContains: "'all' is not a valid region",
		},
		{
			name:        "invalid with s3 bucket regions all",
			config:      NoclickopsConfig{s3Bucket: "my-bucket", s3BucketRegion: "us-east-1", regions: "all"},
			wantErr:     true,
			errContains: "'all' is not a valid region",
		},
		{
			name:            "valid with s3 bucket and specific regions",
			config:          NoclickopsConfig{s3Bucket: "my-bucket", s3BucketRegion: "us-east-1", regions: "us-east-1,eu-west-1"},
			wantErr:         false,
			wantRegionsList: []string{"us-east-1", "eu-west-1"},
		},
		{
			name:            "multiple valid regions populates regionsList",
			config:          NoclickopsConfig{stateFile: "state.tfstate", regions: "us-east-1,eu-west-1"},
			wantErr:         false,
			wantRegionsList: []string{"us-east-1", "eu-west-1"},
		},
		{
			name:            "regions with extra whitespace are trimmed",
			config:          NoclickopsConfig{stateFile: "state.tfstate", regions: " us-east-1 , eu-west-1 "},
			wantErr:         false,
			wantRegionsList: []string{"us-east-1", "eu-west-1"},
		},
		{
			name:            "regions are lowercased",
			config:          NoclickopsConfig{stateFile: "state.tfstate", regions: "US-EAST-1,EU-WEST-1"},
			wantErr:         false,
			wantRegionsList: []string{"us-east-1", "eu-west-1"},
		},
		{
			name:              "single tag is parsed correctly",
			config:            NoclickopsConfig{stateFile: "state.tfstate", regions: "eu-west-1", ignoreTags: "a-tag=a-value"},
			wantErr:           false,
			wantIgnoreTagsMap: map[string][]string{"a-tag": []string{"a-value"}},
		},
		{
			name:              "multiple tags with special characters are parsed correctly",
			config:            NoclickopsConfig{stateFile: "state.tfstate", regions: "eu-west-1", ignoreTags: "a-tag=a-value,another/tag=another.value/with-special-chars"},
			wantErr:           false,
			wantIgnoreTagsMap: map[string][]string{"a-tag": []string{"a-value"}, "another/tag": []string{"another.value/with-special-chars"}},
		},
		{
			name:              "same key multiple times is parsed properly",
			config:            NoclickopsConfig{stateFile: "state.tfstate", regions: "eu-west-1", ignoreTags: "a-tag=a-value,a-tag=b-value,b-tag=c-value"},
			wantErr:           false,
			wantIgnoreTagsMap: map[string][]string{"a-tag": []string{"a-value", "b-value"}, "b-tag": []string{"c-value"}},
		},
		{
			name:        "neither statefile nor s3 bucket",
			config:      NoclickopsConfig{regions: "all"},
			wantErr:     true,
			errContains: "At least one of 's3-bucket' or 'statefile' must be provided",
		},
		{
			name:        "s3 bucket without region",
			config:      NoclickopsConfig{s3Bucket: "my-bucket", regions: "all"},
			wantErr:     true,
			errContains: "s3-bucket-region must be provided if s3-bucket is defined",
		},
		{
			name:        "s3 bucket region set to all",
			config:      NoclickopsConfig{s3Bucket: "my-bucket", s3BucketRegion: "all", regions: "all"},
			wantErr:     true,
			errContains: "'all' is not a valid region",
		},
		{
			name:        "invalid s3 bucket region",
			config:      NoclickopsConfig{s3Bucket: "my-bucket", s3BucketRegion: "not-a-region", regions: "all"},
			wantErr:     true,
			errContains: "'not-a-region' is not a valid region",
		},
		{
			name:        "invalid region in regions list",
			config:      NoclickopsConfig{stateFile: "state.tfstate", regions: "us-east-1,bad-region"},
			wantErr:     true,
			errContains: "'bad-region' is not a valid region",
		},
		{
			name:        "multiple errors reported together",
			config:      NoclickopsConfig{regions: "bad-region"},
			wantErr:     true,
			errContains: "At least one of 's3-bucket' or 'statefile' must be provided",
		},
		{
			name:        "error message includes help hint",
			config:      NoclickopsConfig{regions: "all"},
			wantErr:     true,
			errContains: "Use -h / --help",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validate() error = %q, want it to contain %q", err.Error(), tt.errContains)
				}
				return
			}
			if tt.wantRegionsList != nil {
				if len(tt.config.regionsList) != len(tt.wantRegionsList) {
					t.Errorf("regionsList length = %v, want %v", len(tt.config.regionsList), len(tt.wantRegionsList))
					return
				}
				for _, r := range tt.wantRegionsList {
					if !slices.Contains(tt.config.regionsList, r) {
						t.Errorf("Missing %v from %v", r, tt.config.regionsList)
					}
				}
			}
			if tt.wantIgnoreTagsMap != nil {
				if len(tt.wantIgnoreTagsMap) != len(tt.config.ignoreTagsMap) {
					t.Errorf("ignoreTagsMap length = %v, want %v", len(tt.config.ignoreTagsMap), len(tt.wantIgnoreTagsMap))
					return
				}
				if diff := cmp.Diff(tt.config.ignoreTagsMap, tt.wantIgnoreTagsMap); diff != "" {
					t.Errorf("expected %v, got %v", tt.wantIgnoreTagsMap, tt.config.ignoreTagsMap)
				}
			}
			// When regions == "all", regionsList should be populated with all known regions
			if tt.config.regions == "all" && !tt.wantErr {
				if len(tt.config.regionsList) == 0 {
					t.Errorf("regionsList should be populated when regions is 'all'")
				}
			}
		})
	}
}
