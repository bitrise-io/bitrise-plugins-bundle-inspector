package util

import "testing"

func TestParseSize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{
			name:    "bytes",
			input:   "100B",
			want:    100,
			wantErr: false,
		},
		{
			name:    "bytes without unit",
			input:   "100",
			want:    100,
			wantErr: false,
		},
		{
			name:    "kilobytes",
			input:   "10KB",
			want:    10 * 1024,
			wantErr: false,
		},
		{
			name:    "megabytes",
			input:   "5MB",
			want:    5 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "gigabytes",
			input:   "2GB",
			want:    2 * 1024 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "decimal megabytes",
			input:   "1.5MB",
			want:    int64(1.5 * 1024 * 1024),
			wantErr: false,
		},
		{
			name:    "lowercase",
			input:   "10mb",
			want:    10 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "with spaces",
			input:   "  10 MB  ",
			want:    10 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "invalid number",
			input:   "abc",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid unit",
			input:   "10TB",
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSize(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
