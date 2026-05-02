package tests_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/MintzyG/fun"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func makeReq(method, path string, body string, params map[string]string) *fun.Request {
	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	} else {
		bodyReader = strings.NewReader("")
	}

	r := httptest.NewRequest(method, path, bodyReader)
	r.Header.Set("Content-Type", "application/json")

	// Simulate path params via context (chi-style fake)
	fun.SetPathParamFunc(func(r *http.Request, key string) string {
		return params[key]
	})

	return fun.From(r)
}

func makeQueryReq(queryString string) *fun.Request {
	r := httptest.NewRequest(http.MethodGet, "/?"+queryString, nil)
	return fun.From(r)
}

// ── Value: String ────────────────────────────────────────────────────────────

func TestValue_String(t *testing.T) {
	req := makeReq("GET", "/", "", map[string]string{"name": "sophia"})
	got := req.Path("name").String()
	if got != "sophia" {
		t.Errorf("expected sophia, got %q", got)
	}
}

func TestValue_StringOr_missing(t *testing.T) {
	req := makeReq("GET", "/", "", map[string]string{})
	got := req.Path("name").StringOr("default")
	if got != "default" {
		t.Errorf("expected default, got %q", got)
	}
}

// ── Value: Int ───────────────────────────────────────────────────────────────

func TestValue_Int_valid(t *testing.T) {
	req := makeQueryReq("page=3")
	n, err := req.Query("page").Int()
	if err != nil || n != 3 {
		t.Errorf("expected 3, got %d %v", n, err)
	}
}

func TestValue_Int_invalid(t *testing.T) {
	req := makeQueryReq("page=abc")
	_, err := req.Query("page").Int()
	if err == nil {
		t.Error("expected parse error, got nil")
	}
}

func TestValue_IntOr_missing(t *testing.T) {
	req := makeQueryReq("")
	got := req.Query("page").IntOr(1)
	if got != 1 {
		t.Errorf("expected fallback 1, got %d", got)
	}
}

// ── Value: Bool ──────────────────────────────────────────────────────────────

func TestValue_Bool_true(t *testing.T) {
	req := makeQueryReq("active=true")
	b, err := req.Query("active").Bool()
	if err != nil || !b {
		t.Errorf("expected true, got %v %v", b, err)
	}
}

func TestValue_Bool_invalid(t *testing.T) {
	req := makeQueryReq("active=maybe")
	_, err := req.Query("active").Bool()
	if err == nil {
		t.Error("expected parse error")
	}
}

// ── Value: UUID ──────────────────────────────────────────────────────────────

func TestValue_UUID_valid(t *testing.T) {
	id := "018f4e2a-1234-7000-8000-abcdef012345"
	req := makeReq("GET", "/", "", map[string]string{"id": id})
	parsed, err := req.Path("id").UUID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id {
		t.Errorf("expected %s, got %s", id, parsed.String())
	}
}

func TestValue_UUID_invalid(t *testing.T) {
	req := makeReq("GET", "/", "", map[string]string{"id": "not-a-uuid"})
	_, err := req.Path("id").UUID()
	if err == nil {
		t.Error("expected parse error")
	}
}

func TestValue_UUID_missing(t *testing.T) {
	req := makeReq("GET", "/", "", map[string]string{})
	_, err := req.Path("id").UUID()
	if err == nil {
		t.Error("expected missing error")
	}
	if err, ok := errors.AsType[fun.MissingParamError](err); ok {
		t.Errorf("expected MissingParamError, got %T", err)
	}
}

// ── Value: Enum ──────────────────────────────────────────────────────────────

func TestValue_Enum_valid(t *testing.T) {
	req := makeQueryReq("status=active")
	got, err := req.Query("status").Enum("active", "archived")
	if err != nil || got != "active" {
		t.Errorf("expected active, got %q %v", got, err)
	}
}

func TestValue_Enum_invalid(t *testing.T) {
	req := makeQueryReq("status=deleted")
	_, err := req.Query("status").Enum("active", "archived")
	if err == nil {
		t.Error("expected parse error")
	}
}

func TestValue_EnumOr(t *testing.T) {
	req := makeQueryReq("status=deleted")
	got := req.Query("status").EnumOr("active", "active", "archived")
	if got != "active" {
		t.Errorf("expected fallback active, got %q", got)
	}
}

// ── Value: StripBearer ───────────────────────────────────────────────────────

func TestValue_StripBearer(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer mytoken123")
	req := fun.From(r)
	got := req.Header("Authorization").StripBearer()
	if got != "mytoken123" {
		t.Errorf("expected mytoken123, got %q", got)
	}
}

func TestValue_StripBearer_noPrefix(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "mytoken123")
	req := fun.From(r)
	got := req.Header("Authorization").StripBearer()
	if got != "mytoken123" {
		t.Errorf("expected mytoken123 unchanged, got %q", got)
	}
}

// ── Value: Required ──────────────────────────────────────────────────────────

func TestValue_Required_present(t *testing.T) {
	req := makeQueryReq("name=sophia")
	if err := req.Query("name").Required(); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValue_Required_missing(t *testing.T) {
	req := makeQueryReq("")
	if err := req.Query("name").Required(); err == nil {
		t.Error("expected error for missing required param")
	}
}

// ── QueryAll ─────────────────────────────────────────────────────────────────

func TestQueryAll(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?tag=go&tag=api&tag=rest", nil)
	req := fun.From(r)
	tags := req.QueryAll("tag")
	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(tags))
	}
}

// ── Body ─────────────────────────────────────────────────────────────────────

func TestBody_Into_valid(t *testing.T) {
	type Input struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	body := `{"name":"sophia","age":21}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	req := fun.From(r)

	var input Input
	if err := req.Body().Into(&input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.Name != "sophia" || input.Age != 21 {
		t.Errorf("unexpected input: %+v", input)
	}
}

func TestBody_Into_invalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{bad json}"))
	req := fun.From(r)
	var input struct{ Name string }
	if err := req.Body().Into(&input); err == nil {
		t.Error("expected decode error")
	}
}

func TestBody_Into_unknownField(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"x","unknown":"y"}`))
	req := fun.From(r)
	var input struct {
		Name string `json:"name"`
	}
	// Into uses DisallowUnknownFields
	if err := req.Body().Into(&input); err != nil {
		t.Error("expected no error for unknown field")
	}
}

func TestBody_IntoLenient_unknownField(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"x","unknown":"y"}`))
	req := fun.From(r)
	var input struct {
		Name string `json:"name"`
	}
	if err := req.Body().Into(&input); err != nil {
		t.Errorf("expected no error with lenient decode, got %v", err)
	}
}

func TestBody_Limit(t *testing.T) {
	big := strings.Repeat("a", 1000)
	body := `{"name":"` + big + `"}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req := fun.From(r)
	var input struct {
		Name string `json:"name"`
	}
	// 10 byte limit should cause truncated read → decode error
	if err := req.Body().Limit(10).Into(&input); err == nil {
		t.Error("expected error due to size limit")
	}
}

func TestBody_Text(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello world"))
	req := fun.From(r)
	txt, err := req.Body().Text()
	if err != nil || txt != "hello world" {
		t.Errorf("expected 'hello world', got %q %v", txt, err)
	}
}

// ── Collector ─────────────────────────────────────────────────────────────────

func TestCollector_noErrors(t *testing.T) {
	req := makeReq("GET", "/", "", map[string]string{"id": "018f4e2a-1234-7000-8000-abcdef012345"})
	c := fun.Collect(req)
	_ = c.Path("id").UUID()
	if c.HasErrors() {
		t.Error("expected no errors")
	}
	if c.Fail() != nil {
		t.Error("expected Fail() to return nil")
	}
}

func TestCollector_withErrors(t *testing.T) {
	req := makeReq("GET", "/", "", map[string]string{"id": "not-a-uuid"})
	c := fun.Collect(req)
	_ = c.Path("id").UUID()
	_ = c.Query("page").Int() // missing
	if !c.HasErrors() {
		t.Error("expected errors")
	}
	appErr := c.Fail()
	if appErr == nil {
		t.Fatal("expected non-nil AppError")
	}
	if len(appErr.Fields) != 2 {
		t.Errorf("expected 2 field errors, got %d", len(appErr.Fields))
	}
}

func TestCollector_IntOr_noError(t *testing.T) {
	req := makeQueryReq("page=abc") // invalid but using OrFallback
	c := fun.Collect(req)
	page := c.Query("page").IntOr(1)
	if page != 1 {
		t.Errorf("expected fallback 1, got %d", page)
	}
	// IntOr should not record an error
	if c.HasErrors() {
		t.Error("IntOr should not collect errors")
	}
}

func TestCollector_Enum_recordsError(t *testing.T) {
	req := makeQueryReq("status=invalid")
	c := fun.Collect(req)
	_ = c.Query("status").Enum("active", "archived")
	if !c.HasErrors() {
		t.Error("expected error for invalid enum")
	}
}

func TestCollector_EnumOr_noError(t *testing.T) {
	req := makeQueryReq("status=invalid")
	c := fun.Collect(req)
	got := c.Query("status").EnumOr("active", "active", "archived")
	if got != "active" {
		t.Errorf("expected fallback active, got %q", got)
	}
	if c.HasErrors() {
		t.Error("EnumOr should not collect errors")
	}
}

// ── Bind (via request.From + request.Body) ───────────────────────────────────

func TestBind_validBody(t *testing.T) {
	body := `{"name":"sophia","age":21}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	req := fun.From(r)

	var input struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	if err := req.Body().Into(&input); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if input.Name != "sophia" {
		t.Errorf("bad decode: %+v", input)
	}
}

// ── Form ─────────────────────────────────────────────────────────────────────

func TestBody_Form(t *testing.T) {
	form := url.Values{"username": {"sophia"}, "age": {"21"}}
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req := fun.From(r)

	if err := req.Body().Form(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	username := req.Body().FormValue("username").String()
	if username != "sophia" {
		t.Errorf("expected sophia, got %q", username)
	}
}
