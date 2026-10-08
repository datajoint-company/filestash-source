package common

import "testing"

// The helpers exist because Configuration.Bool() and .Int() type-switch on real
// bool/numeric values and yield false/0 for a string, so a string-typed default
// silently disables the setting it is meant to configure. These tests pin the
// conversion and, as importantly, the fallback: an unparseable or non-positive
// value must leave the stock default in place rather than resolve to the zero
// value, which for a duration means "time out at once" rather than "no limit".

func TestDefaultValueBool(t *testing.T) {
	for _, tc := range []struct {
		env  string
		dval bool
		want bool
	}{
		{"", true, true},
		{"", false, false},
		{"true", false, true},
		{"TRUE", false, true},
		{" on ", false, true},
		{"1", false, true},
		{"yes", false, true},
		{"false", true, false},
		{"off", true, false},
		{"0", true, false},
		{"no", true, false},
		// Unrecognised: keep dval. share.enable defaults to true, so this is the
		// case where failing to the zero value would silently enable sharing.
		{"disabled", true, true},
		{"flase", true, true},
		{"flase", false, false},
	} {
		t.Setenv("FS_TEST_BOOL", tc.env)
		if got := defaultValueBool(tc.dval, "FS_TEST_BOOL"); got != tc.want {
			t.Errorf("defaultValueBool(%v, %q=%q) = %v, want %v", tc.dval, "FS_TEST_BOOL", tc.env, got, tc.want)
		}
	}
}

func TestDefaultValueInt(t *testing.T) {
	for _, tc := range []struct {
		env  string
		dval int
		want int
	}{
		{"", 60, 60},
		{"600", 60, 600},
		{" 64 ", 0, 64},
		// 0 stays 0 where dval is already 0 -- that is upload_chunk_size, for which
		// 0 is the documented "no chunking" value.
		{"0", 0, 0},
		// But where dval is positive the setting is a duration, not a switch:
		// zip_timeout feeds time.Duration(...)*time.Second, so 0 or a negative
		// number would make every zip time out immediately.
		{"0", 60, 60},
		{"-1", 60, 60},
		{"-5", 0, -5},
		// Unparseable: keep dval rather than resolving to 0.
		{"sixty", 60, 60},
		{"64MB", 0, 0},
	} {
		t.Setenv("FS_TEST_INT", tc.env)
		if got := DefaultValueInt(tc.dval, "FS_TEST_INT"); got != tc.want {
			t.Errorf("DefaultValueInt(%d, %q=%q) = %d, want %d", tc.dval, "FS_TEST_INT", tc.env, got, tc.want)
		}
	}
}

func TestDefaultValueEnum(t *testing.T) {
	for _, tc := range []struct {
		env  string
		dval string
		opts []string
		want string
	}{
		{"", "grid", filePageViewOpts, "grid"},
		{"list", "grid", filePageViewOpts, "list"},
		{" list ", "grid", filePageViewOpts, "list"},
		{"date", "type", filePageSortOpts, "date"},
		// Outside opts: keep dval. The frontend treats any non-empty string as the
		// chosen layout, so passing a typo through renders neither layout.
		{"lst", "grid", filePageViewOpts, "grid"},
		{"List", "grid", filePageViewOpts, "grid"},
		{"size", "type", filePageSortOpts, "type"},
	} {
		t.Setenv("FS_TEST_ENUM", tc.env)
		if got := defaultValueEnum(tc.dval, "FS_TEST_ENUM", tc.opts); got != tc.want {
			t.Errorf("defaultValueEnum(%q, %q=%q, %v) = %q, want %q", tc.dval, "FS_TEST_ENUM", tc.env, tc.opts, got, tc.want)
		}
	}
}

// Opts is what the Admin Console offers; the enum helper is what the environment is
// checked against. If they ever diverge, a value the console accepts would be
// rejected from the environment, or the reverse.
func TestSelectOptsMatchValidation(t *testing.T) {
	cfg := NewConfiguration()
	for _, tc := range []struct {
		path string
		opts []string
	}{
		{"general.filepage_default_view", filePageViewOpts},
		{"general.filepage_default_sort", filePageSortOpts},
	} {
		var advertised []string
		cfg.Get(tc.path).Schema(func(f *FormElement) *FormElement {
			if f != nil {
				advertised = f.Opts
			}
			return f
		})
		if len(advertised) != len(tc.opts) {
			t.Fatalf("%s advertises %v but is validated against %v", tc.path, advertised, tc.opts)
		}
		for i := range advertised {
			if advertised[i] != tc.opts[i] {
				t.Errorf("%s advertises %v but is validated against %v", tc.path, advertised, tc.opts)
				break
			}
		}
	}
}

// The share feature is off unless a deployment asks for it. A share carries the
// creator's session, so the link holder acts as the creator and bypasses platform
// RBAC; a default of true would make every chart that forgets the setting unsafe.
// Pinned here because an upstream merge would otherwise flip it back silently.
func TestShareDisabledByDefault(t *testing.T) {
	shareDefault := func() interface{} {
		var d interface{}
		cfg := NewConfiguration()
		cfg.Get("features.share.enable").Schema(func(f *FormElement) *FormElement {
			if f != nil {
				d = f.Default
			}
			return f
		})
		return d
	}

	t.Setenv("FS_SHARE_ENABLE", "")
	if got := shareDefault(); got != false {
		t.Errorf("features.share.enable defaults to %v, want false", got)
	}

	t.Setenv("FS_SHARE_ENABLE", "true")
	if got := shareDefault(); got != true {
		t.Errorf("FS_SHARE_ENABLE=true gives %v, want true", got)
	}
}
