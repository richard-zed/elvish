package edit

import "testing"

func TestFilterSpan(t *testing.T) {
	tests := []struct {
		buffer     string
		wantSeg    string // expected filter segment
		wantPrefix string // expected prefix before the span
	}{
		{"just", "just", ""},
		{"ls; fo", "fo", "ls; "},
		{"ls | fo", "fo", "ls | "},
		{"ls && fo", "fo", "ls && "},
		{"ls\nfo", "fo", "ls\n"},
		{"ls |", "", "ls |"},
		{"ls;", "", "ls;"},
		{"", "", ""},
		{"ls; fo bar", "fo bar", "ls; "},
		{`echo "a; b"`, `echo "a; b"`, ""},
		{`echo 'a | b'`, `echo 'a | b'`, ""},
		{`echo $(a; b)`, `echo $(a; b)`, ""},
		{"ls;  fo", "fo", "ls;  "},
	}
	for _, tc := range tests {
		t.Run(tc.buffer, func(t *testing.T) {
			start, end := filterSpan(tc.buffer)
			if got := tc.buffer[start:end]; got != tc.wantSeg {
				t.Errorf("segment = %q, want %q", got, tc.wantSeg)
			}
			if got := tc.buffer[:start]; got != tc.wantPrefix {
				t.Errorf("prefix = %q, want %q", got, tc.wantPrefix)
			}
		})
	}
}
