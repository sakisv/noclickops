package main

import (
	"strings"
	"testing"
)

func TestIsValidRegion(t *testing.T) {
	tests := []struct {
		region string
		valid  bool
	}{
		{"us-east-1", true},
		{"eu-west-1", true},
		{"all", true},
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
		name            string
		opts            options
		wantErr         bool
		errContains     string
		wantRegionsList []string
	}{
		{
			name:            "valid with statefile only regions all",
			opts:            options{stateFile: "state.tfstate", regions: "all"},
			wantErr:         false,
			wantRegionsList: []string{"eu-west-1", "us-east-1", "us-east-2"},
		},
		{
			name:            "valid with s3 bucket",
			opts:            options{s3Bucket: "my-bucket", s3BucketRegion: "us-east-1", regions: "all"},
			wantErr:         false,
			wantRegionsList: []string{"eu-west-1", "us-east-1", "us-east-2"},
		},
		{
			name:            "valid with s3 bucket and specific regions",
			opts:            options{s3Bucket: "my-bucket", s3BucketRegion: "us-east-1", regions: "us-east-1,eu-west-1"},
			wantErr:         false,
			wantRegionsList: []string{"us-east-1", "eu-west-1"},
		},
		{
			name:            "multiple valid regions populates regionsList",
			opts:            options{stateFile: "state.tfstate", regions: "us-east-1,eu-west-1"},
			wantErr:         false,
			wantRegionsList: []string{"us-east-1", "eu-west-1"},
		},
		{
			name:            "regions with extra whitespace are trimmed",
			opts:            options{stateFile: "state.tfstate", regions: " us-east-1 , eu-west-1 "},
			wantErr:         false,
			wantRegionsList: []string{"us-east-1", "eu-west-1"},
		},
		{
			name:            "regions are lowercased",
			opts:            options{stateFile: "state.tfstate", regions: "US-EAST-1,EU-WEST-1"},
			wantErr:         false,
			wantRegionsList: []string{"us-east-1", "eu-west-1"},
		},
		{
			name:        "neither statefile nor s3 bucket",
			opts:        options{regions: "all"},
			wantErr:     true,
			errContains: "At least one of 's3-bucket' or 'statefile' must be provided",
		},
		{
			name:        "s3 bucket without region",
			opts:        options{s3Bucket: "my-bucket", regions: "all"},
			wantErr:     true,
			errContains: "s3-bucket-region must be provided if s3-bucket is defined",
		},
		{
			name:        "s3 bucket region set to all",
			opts:        options{s3Bucket: "my-bucket", s3BucketRegion: "all", regions: "all"},
			wantErr:     true,
			errContains: "s3-bucket-region cannot be 'all'",
		},
		{
			name:        "invalid s3 bucket region",
			opts:        options{s3Bucket: "my-bucket", s3BucketRegion: "not-a-region", regions: "all"},
			wantErr:     true,
			errContains: "'not-a-region' is not a valid region",
		},
		{
			name:        "invalid region in regions list",
			opts:        options{stateFile: "state.tfstate", regions: "us-east-1,bad-region"},
			wantErr:     true,
			errContains: "'bad-region' is not a valid region",
		},
		{
			name:        "multiple errors reported together",
			opts:        options{regions: "bad-region"},
			wantErr:     true,
			errContains: "At least one of 's3-bucket' or 'statefile' must be provided",
		},
		{
			name:        "error message includes help hint",
			opts:        options{regions: "all"},
			wantErr:     true,
			errContains: "Use -h / --help",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.validate()
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
				if len(tt.opts.regionsList) != len(tt.wantRegionsList) {
					t.Errorf("regionsList length = %v, want %v", len(tt.opts.regionsList), len(tt.wantRegionsList))
					return
				}
				for i, r := range tt.wantRegionsList {
					if tt.opts.regionsList[i] != r {
						t.Errorf("regionsList[%d] = %q, want %q", i, tt.opts.regionsList[i], r)
					}
				}
			}
			// When regions == "all", regionsList should be populated with all known regions
			if tt.opts.regions == "all" && !tt.wantErr {
				if len(tt.opts.regionsList) == 0 {
					t.Errorf("regionsList should be populated when regions is 'all'")
				}
			}
		})
	}
}
