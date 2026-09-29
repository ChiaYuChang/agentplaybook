package version

import "testing"

func TestRelease(t *testing.T) {
	want, ok := parseRelease(embedded)
	if !ok {
		t.Fatal("embedded release version is invalid")
	}
	if got := Release(); got != want {
		t.Fatalf("Release() = %q, want %q", got, want)
	}
}

func TestParseRelease(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "zero components", input: "v0.0.0\n", want: "v0.0.0"},
		{name: "multiple digits", input: "v10.20.30\n", want: "v10.20.30"},
		{name: "missing prefix", input: "1.2.3\n", wantErr: true},
		{name: "missing patch", input: "v1.2\n", wantErr: true},
		{name: "missing final newline", input: "v1.2.3", wantErr: true},
		{name: "extra newline", input: "v1.2.3\n\n", wantErr: true},
		{name: "leading zero major", input: "v01.2.3\n", wantErr: true},
		{name: "leading zero minor", input: "v1.02.3\n", wantErr: true},
		{name: "leading zero patch", input: "v1.2.03\n", wantErr: true},
		{name: "carriage return", input: "v1.2.3\r\n", wantErr: true},
		{name: "whitespace", input: " v1.2.3\n", wantErr: true},
		{name: "suffix", input: "v1.2.3-dev\n", wantErr: true},
		{name: "extra data", input: "v1.2.3\nextra\n", wantErr: true},
		{name: "embedded NUL", input: "v1.2.3\x00\n", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := parseRelease(test.input)
			if test.wantErr {
				if ok {
					t.Fatalf("parseRelease(%q) = %q, want invalid", test.input, got)
				}
				return
			}
			if !ok {
				t.Fatalf("parseRelease(%q) was invalid", test.input)
			}
			if got != test.want {
				t.Errorf("parseRelease(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestReleaseValueFailsClosed(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("releaseValue accepted invalid embedded data")
		}
	}()

	releaseValue("v1.2.3")
}
