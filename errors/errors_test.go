package fe_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	errs "github.com/MintzyG/fun/errors"
)

// ── Problem serialization ────────────────────────────────────────────────────

func TestMarshal_StandardFields(t *testing.T) {
	p := errs.New("https://example.com/probs/out-of-credit").
		Title("Insufficient credit.").
		Detail("Your balance is 30, but that costs 50.").
		Build(403)

	obj := mustMarshal(t, p)

	assertStr(t, obj, "type", "https://example.com/probs/out-of-credit")
	assertStr(t, obj, "title", "Insufficient credit.")
	assertStr(t, obj, "detail", "Your balance is 30, but that costs 50.")
	assertFloat(t, obj, "status", 403)
}

func TestMarshal_ExtensionsAreFlat(t *testing.T) {
	p := errs.New("https://example.com/probs/out-of-credit").
		Title("Insufficient credit.").
		With("balance", 30).
		With("accounts", []string{"/account/12345", "/account/67890"}).
		Build(403)

	obj := mustMarshal(t, p)

	if _, ok := obj["balance"]; !ok {
		t.Error("expected 'balance' at top level")
	}
	if _, ok := obj["accounts"]; !ok {
		t.Error("expected 'accounts' at top level")
	}
	if _, ok := obj["extensions"]; ok {
		t.Error("must not have an 'extensions' wrapper key")
	}
}

func TestMarshal_ReservedNamesIgnored(t *testing.T) {
	p := errs.New(errs.AboutBlank).
		With("type", "attacker-controlled").
		With("status", 200).
		Build(404)

	obj := mustMarshal(t, p)

	assertStr(t, obj, "type", errs.AboutBlank)
	assertFloat(t, obj, "status", 404)
}

func TestMarshal_DefaultTypeIsAboutBlank(t *testing.T) {
	p := errs.From(fmt.Errorf("something broke")).Build(500)
	obj := mustMarshal(t, p)
	assertStr(t, obj, "type", "about:blank")
}

// ── Generic helpers ──────────────────────────────────────────────────────────

func TestGeneric_ReturnProblem(t *testing.T) {
	cases := []struct {
		p      *errs.Problem
		status int
	}{
		{errs.BadRequest("bad input"), 400},
		{errs.Unauthorized("go away"), 401},
		{errs.Forbidden("nope"), 403},
		{errs.NotFound("missing"), 404},
		{errs.Conflict("already exists"), 409},
		{errs.Internal("boom"), 500},
	}
	for _, c := range cases {
		if c.p.Status() != c.status {
			t.Errorf("expected %d, got %d: %s", c.status, c.p.Status(), c.p.Error())
		}
	}
}

func TestGeneric_DefaultTitle(t *testing.T) {
	p := errs.NotFound()
	if p.Title() != "Not Found" {
		t.Errorf("default title: got %q", p.Title())
	}
}

func TestGeneric_CustomTitle(t *testing.T) {
	p := errs.NotFound("Subject not found in context")
	if p.Title() != "Subject not found in context" {
		t.Errorf("custom title: got %q", p.Title())
	}
}

// ── NewXxx builder helpers ───────────────────────────────────────────────────

func TestNew_PreloadsStatusAndTitle(t *testing.T) {
	p := errs.NewNotFound().Err()
	if p.Status() != 404 {
		t.Errorf("status: got %d, want 404", p.Status())
	}
	if p.Title() != "Not Found" {
		t.Errorf("title: got %q, want 'Not Found'", p.Title())
	}
}

func TestNew_CanOverrideType(t *testing.T) {
	p := errs.NewNotFound().
		Type("https://docs.example.com/problems/missing-subject").
		Detail("subject was not found in request context").
		Err()

	if p.Status() != 404 {
		t.Errorf("status: got %d, want 404", p.Status())
	}
	if p.Type() != "https://docs.example.com/problems/missing-subject" {
		t.Errorf("type: got %q", p.Type())
	}
	if p.Detail() != "subject was not found in request context" {
		t.Errorf("detail: got %q", p.Detail())
	}
}

func TestNew_WithExtensions(t *testing.T) {
	p := errs.NewForbidden().
		With("required_role", "admin").
		Err()

	obj := mustMarshal(t, p)
	if obj["required_role"] != "admin" {
		t.Errorf("extension: got %v", obj["required_role"])
	}
}

// ── error interface ──────────────────────────────────────────────────────────

func TestProblem_ErrorString(t *testing.T) {
	p := errs.New(errs.AboutBlank).
		Title("Not Found").
		Detail("thing is gone").
		Build(404)

	got := p.Error()
	if !strings.Contains(got, "404") {
		t.Errorf("Error() missing status: %q", got)
	}
	if !strings.Contains(got, "Not Found") {
		t.Errorf("Error() missing title: %q", got)
	}
}

func TestProblem_Unwrap(t *testing.T) {
	inner := fmt.Errorf("root cause")
	p := errs.From(inner).Build(400)
	if !errors.Is(inner, p.Unwrap()) {
		t.Error("Unwrap should return the original cause")
	}
}

// ── Validation ───────────────────────────────────────────────────────────────

func TestValidation_ErrorsExtension(t *testing.T) {
	p := errs.Validation("https://example.net/validation-error").
		Title("Your request is not valid.").
		Field("body", "age", "must be a positive integer").
		Field("body", "profile/color", "must be 'green', 'red' or 'blue'").
		Build()

	if p.Status() != 422 {
		t.Errorf("status: got %d, want 422", p.Status())
	}
	if p.Extension("errors") == nil {
		t.Fatal("expected 'errors' extension")
	}

	obj := mustMarshal(t, p)
	errsExt, ok := obj["errors"].([]any)
	if !ok || len(errsExt) != 2 {
		t.Fatalf("expected 2 validation errors, got: %v", obj["errors"])
	}

	first := errsExt[0].(map[string]any)
	if first["pointer"] != "#/body/age" {
		t.Errorf("pointer: got %q", first["pointer"])
	}
	if first["detail"] != "must be a positive integer" {
		t.Errorf("detail: got %q", first["detail"])
	}
}

func TestValidation_DefaultTitle(t *testing.T) {
	p := errs.Validation("https://example.net/validation-error").
		Add("#/name", "is required").
		Build()

	if p.Title() != "Your request is not valid." {
		t.Errorf("default title: got %q", p.Title())
	}
}

func TestValidation_HasErrors(t *testing.T) {
	vb := errs.Validation("https://example.net/validation-error")
	if vb.HasErrors() {
		t.Error("should have no errors initially")
	}
	vb.Add("#/name", "is required")
	if !vb.HasErrors() {
		t.Error("should have errors after Add")
	}
}

// ── Resolve ──────────────────────────────────────────────────────────────────

func TestResolve_PassthroughProblem(t *testing.T) {
	original := errs.NotFound("thing missing")
	resolved := errs.Resolve(original)
	if !errors.Is(resolved, original) {
		t.Error("Resolve should return the same *Problem unchanged")
	}
}

func TestResolve_Mapper(t *testing.T) {
	errs.RegisterMapper(func(err error) *errs.Problem {
		if err.Error() == "thing not found" {
			return errs.NotFound("thing not found")
		}
		return nil
	})
	defer errs.ResetMapper()

	p := errs.Resolve(fmt.Errorf("thing not found"))
	if p.Status() != 404 {
		t.Errorf("mapped status: got %d, want 404", p.Status())
	}
}

func TestResolve_NoMapper_Returns500(t *testing.T) {
	errs.ResetMapper()
	p := errs.Resolve(fmt.Errorf("some unmapped error"))
	if p.Status() != 500 {
		t.Errorf("unmapped status: got %d, want 500", p.Status())
	}
}

func TestResolve_Nil_Returns500(t *testing.T) {
	p := errs.Resolve(nil)
	if p.Status() != 500 {
		t.Errorf("nil error status: got %d, want 500", p.Status())
	}
}

// ── Debug ────────────────────────────────────────────────────────────────────

func TestDebug_DisabledByDefault(t *testing.T) {
	p := errs.NewInternal().Debug(fmt.Errorf("raw")).Err()
	if p.Extension("debug") != nil {
		t.Error("debug extension must not appear when debug is disabled")
	}
}

func TestDebug_Enabled(t *testing.T) {
	errs.SetDebug(true)
	defer errs.SetDebug(false)

	p := errs.NewInternal().Debug(fmt.Errorf("raw cause")).Err()
	if p.Extension("debug") == nil {
		t.Error("debug extension must appear when debug is enabled")
	}
}

// ── test helpers ─────────────────────────────────────────────────────────────

func mustMarshal(t *testing.T, p *errs.Problem) map[string]any {
	t.Helper()
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return obj
}

func assertStr(t *testing.T, obj map[string]any, key, want string) {
	t.Helper()
	got, ok := obj[key].(string)
	if !ok {
		t.Errorf("%s: missing or wrong type (got %T)", key, obj[key])
		return
	}
	if got != want {
		t.Errorf("%s: got %q, want %q", key, got, want)
	}
}

func assertFloat(t *testing.T, obj map[string]any, key string, want float64) {
	t.Helper()
	got, ok := obj[key].(float64)
	if !ok {
		t.Errorf("%s: missing or wrong type (got %T)", key, obj[key])
		return
	}
	if got != want {
		t.Errorf("%s: got %v, want %v", key, got, want)
	}
}
