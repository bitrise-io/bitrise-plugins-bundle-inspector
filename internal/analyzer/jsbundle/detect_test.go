package jsbundle

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestIsJSBundleFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"main.jsbundle", "main.jsbundle", true},
		{"custom name jsbundle", "app.jsbundle", true},
		{"capitalized jsbundle", "MyApp.jsbundle", true},
		{"uppercase JSBUNDLE", "Main.JSBUNDLE", true},
		{"android bundle", "index.android.bundle", true},
		{"android bundle uppercase", "INDEX.ANDROID.BUNDLE", true},
		{"plain js file", "main.js", false},
		{"bundle.js", "bundle.js", false},
		{"dex file", "classes.dex", false},
		{"empty string", "", false},
		{"jsbundle no dot", "jsbundle", false},
		{"ios bundle wrong name", "index.ios.bundle", false},
		{"partial match", "notjsbundle", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsJSBundleFilename(tt.filename)
			if got != tt.want {
				t.Errorf("IsJSBundleFilename(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    BundleFormat
		wantErr bool
	}{
		{
			name: "hermes bytecode",
			data: func() []byte {
				buf := make([]byte, 64)
				binary.LittleEndian.PutUint32(buf[0:4], HermesMagic)
				return buf
			}(),
			want: FormatHermes,
		},
		{
			name: "RAM bundle",
			data: func() []byte {
				buf := make([]byte, 64)
				binary.LittleEndian.PutUint32(buf[0:4], RAMBundleMagic)
				return buf
			}(),
			want: FormatRAMBundle,
		},
		{
			name: "metro bundle with __d(function",
			data: []byte(`var __BUNDLE_START_TIME__=this.nativePerformanceNow?nativePerformanceNow():Date.now(),__DEV__=false,process={env:{NODE_ENV:"production"}};
__d(function(g,r,i,a,m,e,d){var n=r(d[0]);Object.defineProperty(e,"__esModule",{value:!0})`),
			want: FormatMetro,
		},
		{
			name: "metro bundle with var prefix",
			data: []byte(`var __DEV__=false;`),
			want: FormatMetro,
		},
		{
			name: "metro bundle with __d( call",
			data: []byte(`some preamble code here; __d(42, [1, 2, 3]);`),
			want: FormatMetro,
		},
		{
			name:    "unknown binary content",
			data:    []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09},
			want:    FormatUnknown,
		},
		{
			name: "empty input",
			data: []byte{},
			want: FormatUnknown,
		},
		{
			name: "short input (2 bytes)",
			data: []byte{0xAB, 0xCD},
			want: FormatUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			got, err := DetectFormat(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DetectFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
