package bytes

import (
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		input uint64
		want  string
	}{
		{0, "0B"},
		{1, "1B"},
		{999, "999B"},
		{1024, "1.00KB"},
		{1536, "1.50KB"},
		{1048576, "1.00MB"},
		{1073741824, "1.00GB"},
		{1099511627776, "1.00TB"},
		{1125899906842624, "1.00PB"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := Format(tt.input)
			if got != tt.want {
				t.Errorf("Format(%d) = %q, want %q", tt.input, got, tt.want)
			}
			got2 := New().Format(tt.input)
			if got2 != tt.want {
				t.Errorf("New().Format(%d) = %q, want %q", tt.input, got2, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		input   string
		want    uint64
		wantErr bool
	}{
		{"0B", 0, false},
		{"1B", 1, false},
		{"5KB", 5 * KB, false},
		{"5K", 5 * KB, false},
		{"10MB", 10 * MB, false},
		{"10M", 10 * MB, false},
		{"2GB", 2 * GB, false},
		{"2G", 2 * GB, false},
		{"1TB", 1 * TB, false},
		{"1T", 1 * TB, false},
		{"3PB", 3 * PB, false},
		{"3P", 3 * PB, false},
		{"", 0, true},
		{"abc", 0, true},
		{"12XB", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Parse(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if KB != 1024 {
		t.Errorf("KB = %d, want 1024", KB)
	}
	if MB != 1024*1024 {
		t.Errorf("MB = %d, want %d", MB, 1024*1024)
	}
	if GB != 1024*1024*1024 {
		t.Errorf("GB = %d, want %d", GB, 1024*1024*1024)
	}
}
