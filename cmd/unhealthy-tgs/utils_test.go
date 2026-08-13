package main

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

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
		{"ab", "/ab/"},
		{"abc", "/abc/"},

		// Valid transformations.
		{"headh2", "/head/h/2/"},
		{"headz9", "/head/z/9/"},
		{"abch1", "/abc/h/1/"},
		{"abcz9", "/abc/z/9/"},
		{"123h1", "/123/h/1/"},

		// Digit bounds.
		{"headh0", "/headh0/"},
		{"devo1", "/dev/o/1/"},
		{"uatx9", "/uat/x/9/"},

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

func TestEnvProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in          string
		wantEnv     string
		wantProfile string
	}{
		{"", "", ""},
		{":", "", ""},
		{": ", "", ""},
		{" :", "", ""},
		{" : ", "", ""},
		{"::", "", ""},
		{" ::", "", ""},
		{": :", "", ""},
		{" : :", "", ""},
		{" : : ", "", ""},

		// env only
		{"env", "env", ""},
		{" env ", "env", ""},
		{"\tenv\t", "env", ""},
		{"non prod", "non prod", ""},
		{"non  prod", "non  prod", ""},
		{"non\tprod", "non\tprod", ""},
		{" non prod ", "non prod", ""},

		// normal
		{"env:profile", "env", "profile"},
		{" env : profile ", "env", "profile"},
		{"\tenv\t:\tprofile\t", "env", "profile"},

		// spaces inside env
		{"non prod:profile", "non prod", "profile"},
		{"non  prod:profile", "non  prod", "profile"},
		{" non prod : profile ", "non prod", "profile"},
		{"non\tprod:profile", "non\tprod", "profile"},

		// spaces inside profile
		{"env:non prod", "env", "non prod"},
		{"env:non  prod", "env", "non  prod"},
		{"env:non\tprod", "env", "non\tprod"},
		{" env : non prod ", "env", "non prod"},

		// spaces inside both
		{"non prod:dev profile", "non prod", "dev profile"},
		{" non prod : dev profile ", "non prod", "dev profile"},
		{"non\tprod:dev\tprofile", "non\tprod", "dev\tprofile"},

		// empty env
		{":profile", "", "profile"},
		{" : profile ", "", "profile"},
		{" : pro file ", "", "pro file"},
		{"\t:\tprofile\t", "", "profile"},
		{"  :  non prod profile  ", "", "non prod profile"},

		// empty profile
		{"env:", "env", ""},
		{" env : ", "env", ""},
		{"env:   ", "env", ""},
		{"env:\t\t", "env", ""},
		{"non prod :   ", "non prod", ""},

		// multiple separators
		{"env:profile:extra", "env", "profile"},
		{"a:b:c", "a", "b"},
		{"a:b:c:d", "a", "b"},
		{" env : profile : extra", "env", "profile"},
		{" env : pro file : extra", "env", "pro file"},

		// profile may contain leading/trailing spaces before second colon
		{"env: profile   :extra", "env", "profile"},
		{"env:\tprofile\t:extra", "env", "profile"},

		// env is empty, extra colon
		{":profile:extra", "", "profile"},
		{" : profile : extra ", "", "profile"},

		// profile empty before second colon
		{"env::extra", "env", ""},
		{"env: :extra", "env", ""},
		{"env:\t:extra", "env", ""},

		// only separators
		{":::", "", ""},
		{" : : : ", "", ""},
		{"\t:\t:\t", "", ""},

		// tabs
		{"\tenv\t:\tprofile\t", "env", "profile"},
		{"\t env name \t:\t profile name \t", "env name", "profile name"},

		// mixed tabs/spaces
		{" \tenv\t name\t : \tprofile\t name\t ", "env\t name", "profile\t name"},

		// preserve internal whitespace
		{"env   name:profile", "env   name", "profile"},
		{"env:profile   name", "env", "profile   name"},
		{"env\t\tname:profile", "env\t\tname", "profile"},
		{"env:profile\t\tname", "env", "profile\t\tname"},

		// Unicode spaces are NOT trimmed by current implementation
		{"\u00a0env\u00a0", "\u00a0env\u00a0", ""},
		{"\u00a0env\u00a0:\u00a0profile\u00a0", "\u00a0env\u00a0", "\u00a0profile\u00a0"},

		// weird but valid
		{" :", "", ""},
		{"env :profile", "env", "profile"},
		{"env: profile", "env", "profile"},
		{" env:profile ", "env", "profile"},
		{"    env    :    profile    ", "env", "profile"},

		// only whitespace
		{" ", "", ""},
		{"   ", "", ""},
		{"\t", "", ""},
		{" \t ", "", ""},

		// whitespace profile names preserved internally
		{"env:a b c", "env", "a b c"},
		{"env:a  b  c", "env", "a  b  c"},
		{"env:a\tb\tc", "env", "a\tb\tc"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			gotEnv, gotProfile := envProfile(tt.in)
			if gotEnv != tt.wantEnv || gotProfile != tt.wantProfile {
				t.Fatalf(
					"envProfile(%q)=(%q,%q), want (%q,%q)",
					tt.in,
					gotEnv, gotProfile,
					tt.wantEnv, tt.wantProfile,
				)
			}
		})
	}
}

func tag(k, v string) types.Tag {
	return types.Tag{
		Key:   new(k),
		Value: new(v),
	}
}

func TestTagsMatchEnv(t *testing.T) {
	tests := []struct {
		name string
		tags []types.Tag
		env  string
		want bool
	}{
		{
			name: "strict Environment exact match",
			env:  "devo1",
			tags: []types.Tag{
				tag("Environment", "devo1"),
			},
			want: true,
		},
		{
			name: "strict ei:environment exact match",
			env:  "devo1",
			tags: []types.Tag{
				tag("ei:environment", "devo1"),
			},
			want: true,
		},
		{
			name: "strict Environment mismatch",
			env:  "devo1",
			tags: []types.Tag{
				tag("Environment", "devo2"),
			},
			want: false,
		},
		{
			name: "strict comparison is case sensitive",
			env:  "devo1",
			tags: []types.Tag{
				tag("Environment", "DEVO1"),
			},
			want: false,
		},
		{
			name: "non strict ProvisioningSource contains env path",
			env:  "devo1",
			tags: []types.Tag{
				tag("ProvisioningSource", "/dev/o/1/app"),
			},
			want: true,
		},
		{
			name: "non strict ei:provisioning-source contains env path",
			env:  "uatx2",
			tags: []types.Tag{
				tag("ei:provisioning-source", "/uat/x/2/service"),
			},
			want: true,
		},
		{
			name: "non strict ProvisioningSource does not contain env path",
			env:  "devo1",
			tags: []types.Tag{
				tag("ProvisioningSource", "/dev/o/2/app"),
			},
			want: false,
		},
		{
			name: "unknown tag ignored",
			env:  "devo1",
			tags: []types.Tag{
				tag("SomeOtherTag", "devo1"),
			},
			want: false,
		},
		{
			name: "empty tags",
			env:  "devo1",
			tags: nil,
			want: false,
		},
		{
			name: "first tag mismatches second matches",
			env:  "devo1",
			tags: []types.Tag{
				tag("Environment", "uatx2"),
				tag("ProvisioningSource", "/dev/o/1/app"),
			},
			want: true,
		},
		{
			name: "first tag matches second ignored",
			env:  "devo1",
			tags: []types.Tag{
				tag("Environment", "devo1"),
				tag("Environment", "uatx2"),
			},
			want: true,
		},
		{
			name: "mixed unrelated tags and match",
			env:  "devo1",
			tags: []types.Tag{
				tag("Owner", "john"),
				tag("CostCenter", "123"),
				tag("Environment", "devo1"),
			},
			want: true,
		},
		{
			name: "env shorter than 3 chars strict match",
			env:  "ab",
			tags: []types.Tag{
				tag("Environment", "ab"),
			},
			want: true,
		},
		{
			name: "env shorter than 3 chars path match",
			env:  "ab",
			tags: []types.Tag{
				tag("ProvisioningSource", "/ab/"),
			},
			want: true,
		},
		{
			name: "strict tag wins by exact value only",
			env:  "devo1",
			tags: []types.Tag{
				tag("Environment", "/dev/o/1/"),
			},
			want: false,
		},
		{
			name: "non strict tag does not require exact equality",
			env:  "devo1",
			tags: []types.Tag{
				tag("ProvisioningSource", "xxx/dev/o/1/yyy"),
			},
			want: true,
		},
		{
			name: "non strict partial env name does not match",
			env:  "devo1",
			tags: []types.Tag{
				tag("ProvisioningSource", "/devo1/app"),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tagsMatchEnv(eiTags, tc.tags, tc.env)
			if got != tc.want {
				t.Fatalf("got=%v want=%v", got, tc.want)
			}
		})
	}
}

// TODO
// func FuzzStrictEnvironmentMatch(f *testing.F) {
// 	f.Add("devo1")
// 	f.Fuzz(func(t *testing.T, env string) {
// 		tag := types.Tag{
// 			Key:   new("Environment"),
// 			Value: &env,
// 		}
// 		if !tagsMatchEnv(eiTags, []types.Tag{tag}, env) {
// 			t.Fatalf("exact Environment match must return true")
// 		}
// 	})
// }

// func FuzzUnknownTagNeverMatches(f *testing.F) {
// 	f.Add("devo1", "UnknownTag", "devo1")
// 	f.Fuzz(func(t *testing.T, env, key, value string) {
// 		switch key {
// 		case "Environment",
// 			"ei:environment",
// 			"ProvisioningSource",
// 			"ei:provisioning-source":
// 			return
// 		}

// 		tag := types.Tag{
// 			Key:   &key,
// 			Value: &value,
// 		}

// 		if got := tagsMatchEnv(eiTags, []types.Tag{tag}, env); got {
// 			t.Fatalf(
// 				"unknown tag matched: key=%q value=%q env=%q",
// 				key,
// 				value,
// 				env,
// 			)
// 		}
// 	})
// }

// func FuzzProvisioningSourceMatch(f *testing.F) {
// 	f.Add("devo1")
// 	f.Fuzz(func(t *testing.T, env string) {
// 		path := env2path(env)
// 		tag := types.Tag{
// 			Key:   new("ProvisioningSource"),
// 			Value: new("prefix" + path + "suffix"),
// 		}
// 		if !tagsMatchEnv(eiTags, []types.Tag{tag}, env) {
// 			t.Fatalf("expected match")
// 		}
// 	})
// }
