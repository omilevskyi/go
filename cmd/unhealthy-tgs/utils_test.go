package main

import "testing"

func TestEnv2Path(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		// Empty and short inputs.
		{"", "//"},
		{"a", "/a/"},
		{"h", "/h/"},
		{"1", "/1/"},
		{"h2", "/h2/"},
		{"z9", "/z9/"},

		// Valid transformations.
		{"headh2", "/head/h/2/"},
		{"headz9", "/head/z/9/"},
		{"abch1", "/abc/h/1/"},
		{"abcz9", "/abc/z/9/"},
		{"123h1", "/123/h/1/"},

		// Digit bounds.
		{"headh0", "/headh0/"},
		{"headh1", "/head/h/1/"},
		{"headh9", "/head/h/9/"},

		// Letter bounds.
		{"headg1", "/headg1/"},
		{"headh1", "/head/h/1/"},
		{"headz1", "/head/z/1/"},

		// Uppercase letters must not match.
		{"headH1", "/headH1/"},
		{"headZ9", "/headZ9/"},

		// Last character is not a digit.
		{"headhx", "/headhx/"},
		{"headzz", "/headzz/"},
		{"headh ", "/headh /"},

		// Penultimate character is not in h-z.
		{"heada1", "/heada1/"},
		{"headb9", "/headb9/"},
		{"head01", "/head01/"},

		// Longer strings.
		{"productionh2", "/production/h/2/"},
		{"productionz9", "/production/z/9/"},
		{"productiong9", "/productiong9/"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			if got := env2path(tt.in); got != tt.want {
				t.Fatalf("env2path(%q)=%q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestEnv2Path_Bounds(t *testing.T) {
	t.Parallel()

	for letter := byte('a'); letter <= 'z'; letter++ {
		for digit := byte('0'); digit <= '9'; digit++ {
			in := "head" + string(letter) + string(digit)

			got := env2path(in)

			shouldMatch :=
				e2pStLet <= letter && letter <= e2pFinLet &&
					e2pStDgt <= digit && digit <= e2pFinDgt

			if shouldMatch {
				want := "/head/" + string(letter) + "/" + string(digit) + "/"
				if got != want {
					t.Fatalf("env2path(%q)=%q, want %q", in, got, want)
				}
			} else {
				want := "/" + in + "/"
				if got != want {
					t.Fatalf("env2path(%q)=%q, want %q", in, got, want)
				}
			}
		}
	}
}
