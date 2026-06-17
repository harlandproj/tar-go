package filters

import "testing"

func TestTransformBasic(t *testing.T) {
	tr, err := NewTransform("s/foo/bar/")
	if err != nil {
		t.Fatal(err)
	}
	result := tr.Apply("/path/foo/file")
	if result != "/path/bar/file" {
		t.Errorf("got %q, want %q", result, "/path/bar/file")
	}
}

func TestTransformGlobal(t *testing.T) {
	tr, err := NewTransform("s/a/X/g")
	if err != nil {
		t.Fatal(err)
	}
	result := tr.Apply("banana")
	if result != "bXnXnX" {
		t.Errorf("got %q, want %q", result, "bXnXnX")
	}
}

func TestTransformNil(t *testing.T) {
	var tr *Transform
	result := tr.Apply("unchanged")
	if result != "unchanged" {
		t.Errorf("got %q, want %q", result, "unchanged")
	}
}

func TestTransformBackref(t *testing.T) {
	tr, err := NewTransform("s/(foo)-(bar)/\\2-\\1/")
	if err != nil {
		t.Fatal(err)
	}
	result := tr.Apply("foo-bar")
	if result != "bar-foo" {
		t.Errorf("got %q, want %q", result, "bar-foo")
	}
}

func TestTransformCaseInsensitive(t *testing.T) {
	tr, err := NewTransform("s/foo/bar/i")
	if err != nil {
		t.Fatal(err)
	}
	result := tr.Apply("FOObar")
	if result != "barbar" {
		t.Errorf("got %q, want %q", result, "barbar")
	}
}

func TestTransformNoMatch(t *testing.T) {
	tr, err := NewTransform("s/foo/bar/")
	if err != nil {
		t.Fatal(err)
	}
	result := tr.Apply("nothing")
	if result != "nothing" {
		t.Errorf("got %q, want %q", result, "nothing")
	}
}

func TestTransformMultipleDelim(t *testing.T) {
	tr, err := NewTransform("s|/|-|g")
	if err != nil {
		t.Fatal(err)
	}
	result := tr.Apply("/a/b/c")
	if result != "-a-b-c" {
		t.Errorf("got %q, want %q", result, "-a-b-c")
	}
}

func TestInvalidTransform(t *testing.T) {
	_, err := NewTransform("bad")
	if err == nil {
		t.Error("expected error for invalid transform")
	}
	_, err = NewTransform("")
	if err == nil {
		t.Error("expected error for empty transform")
	}
}
