package tests_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/MintzyG/fun"
	"github.com/MintzyG/fun/bind"
	"github.com/go-playground/validator/v10"
)

func init() {
	v := validator.New()
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return f.Name
		}
		return name
	})
	_ = v.RegisterValidation("passwd", func(fl validator.FieldLevel) bool {
		p := fl.Field().String()
		var hasUpper, hasNum, hasSym bool
		for _, c := range p {
			switch {
			case c >= 'A' && c <= 'Z':
				hasUpper = true
			case c >= '0' && c <= '9':
				hasNum = true
			case strings.ContainsRune(`!@#$%^&*()_+-=[]{}|;':",.<>?/`, c):
				hasSym = true
			}
		}
		return hasUpper && hasNum && hasSym
	})
	bind.SetValidator(v)
}

func bindReq(body string) *fun.Request {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return fun.From(r)
}

// ── happy path ───────────────────────────────────────────────────────────────

func TestBind_valid(t *testing.T) {
	type Input struct {
		Name  string `json:"name"  validate:"required"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age"   validate:"gte=18"`
	}

	req := bindReq(`{"name":"Sophia","email":"sophia@trie.oh","age":21}`)
	var input Input
	if err := bind.Body(req).Bind(&input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.Name != "Sophia" || input.Email != "sophia@trie.oh" || input.Age != 21 {
		t.Errorf("unexpected decoded value: %+v", input)
	}
}

// ── decode errors ─────────────────────────────────────────────────────────────

func TestBind_invalidJSON(t *testing.T) {
	req := bindReq(`{bad json}`)
	var input struct {
		Name string `json:"name"`
	}
	err := bind.Body(req).Bind(&input)
	if err == nil {
		t.Fatal("expected AppError, got nil")
	}
	if err.Code != "BAD_REQUEST" {
		t.Errorf("expected BAD_REQUEST, got %s", err.Code)
	}
}

func TestBind_unknownField(t *testing.T) {
	req := bindReq(`{"name":"x","ghost":"y"}`)
	var input struct {
		Name string `json:"name" validate:"required"`
	}
	err := bind.Body(req).Bind(&input)
	if err != nil {
		t.Fatal("expected no AppError for unknown field")
	}
}

// ── validation errors ────────────────────────────────────────────────────────

func TestBind_requiredMissing(t *testing.T) {
	type Input struct {
		Name  string `json:"name"  validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}
	req := bindReq(`{}`)
	var input Input
	err := bind.Body(req).Bind(&input)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", err.Code)
	}
	if len(err.Fields) == 0 {
		t.Error("expected field errors")
	}
}

func TestBind_invalidEmail(t *testing.T) {
	type Input struct {
		Email string `json:"email" validate:"required,email"`
	}
	req := bindReq(`{"email":"notanemail"}`)
	var input Input
	err := bind.Body(req).Bind(&input)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Fields[0].Field != "email" {
		t.Errorf("expected field=email, got %q", err.Fields[0].Field)
	}
	if err.Fields[0].Message != "must be a valid email address" {
		t.Errorf("unexpected message: %q", err.Fields[0].Message)
	}
}

func TestBind_minString(t *testing.T) {
	type Input struct {
		Name string `json:"name" validate:"min=5"`
	}
	req := bindReq(`{"name":"ab"}`)
	var input Input
	err := bind.Body(req).Bind(&input)
	if err == nil {
		t.Fatal("expected validation error")
	}
	want := "must be at least 5 characters long"
	if err.Fields[0].Message != want {
		t.Errorf("expected %q, got %q", want, err.Fields[0].Message)
	}
}

func TestBind_minInt(t *testing.T) {
	type Input struct {
		Age int `json:"age" validate:"min=18"`
	}
	req := bindReq(`{"age":10}`)
	var input Input
	err := bind.Body(req).Bind(&input)
	if err == nil {
		t.Fatal("expected validation error")
	}
	want := "must be at least 18"
	if err.Fields[0].Message != want {
		t.Errorf("expected %q, got %q", want, err.Fields[0].Message)
	}
}

func TestBind_oneof(t *testing.T) {
	type Input struct {
		Status string `json:"status" validate:"oneof=active archived draft"`
	}
	req := bindReq(`{"status":"deleted"}`)
	var input Input
	err := bind.Body(req).Bind(&input)
	if err == nil {
		t.Fatal("expected validation error")
	}
	want := "must be one of: active, archived, draft"
	if err.Fields[0].Message != want {
		t.Errorf("expected %q, got %q", want, err.Fields[0].Message)
	}
}

// ── password masking ──────────────────────────────────────────────────────────

func TestBind_passwordNotExposedInError(t *testing.T) {
	type Input struct {
		Email    string `json:"email"    validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}
	req := bindReq(`{"email":"a@b.com","password":"short"}`)
	var input Input
	err := bind.Body(req).Bind(&input)
	if err == nil {
		t.Fatal("expected validation error")
	}
	for _, f := range err.Fields {
		if f.Field == "password" && f.Value != nil {
			t.Errorf("password value should be masked, got %v", f.Value)
		}
	}
}

func TestBind_nonPasswordExposedInError(t *testing.T) {
	type Input struct {
		Email string `json:"email" validate:"required,email"`
	}
	req := bindReq(`{"email":"notanemail"}`)
	var input Input
	err := bind.Body(req).Bind(&input)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Fields[0].Value == nil {
		t.Error("non-password field value should be exposed in error")
	}
}

// ── passwd tag expansion ──────────────────────────────────────────────────────

func TestBind_passwdExpanded_allMissing(t *testing.T) {
	type Input struct {
		Password string `json:"password" validate:"required,passwd"`
	}
	req := bindReq(`{"password":"alllowercase"}`)
	var input Input
	err := bind.Body(req).Bind(&input)
	if err == nil {
		t.Fatal("expected validation error")
	}
	// Should have 3 separate field errors: uppercase, number, symbol
	if len(err.Fields) != 3 {
		t.Errorf("expected 3 passwd field errors, got %d: %+v", len(err.Fields), err.Fields)
	}
}

func TestBind_passwdExpanded_partial(t *testing.T) {
	type Input struct {
		Password string `json:"password" validate:"required,passwd"`
	}
	// Has uppercase and number, missing symbol
	req := bindReq(`{"password":"Uppercase1"}`)
	var input Input
	err := bind.Body(req).Bind(&input)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if len(err.Fields) != 1 {
		t.Errorf("expected 1 passwd field error (symbol), got %d: %+v", len(err.Fields), err.Fields)
	}
	if !strings.Contains(err.Fields[0].Message, "symbol") {
		t.Errorf("expected symbol message, got %q", err.Fields[0].Message)
	}
}

// ── custom tag message ────────────────────────────────────────────────────────

func TestBind_RegisterTagMessage(t *testing.T) {
	bind.RegisterTagMessage("required", "campo obrigatório")
	defer bind.RegisterTagMessage("required", "this field is required") // restore

	type Input struct {
		Name string `json:"name" validate:"required"`
	}
	req := bindReq(`{}`)
	var input Input
	err := bind.Body(req).Bind(&input)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Fields[0].Message != "campo obrigatório" {
		t.Errorf("expected custom message, got %q", err.Fields[0].Message)
	}
}

// ── Limit ─────────────────────────────────────────────────────────────────────

func TestBind_Limit(t *testing.T) {
	big := strings.Repeat("a", 10000)
	req := bindReq(`{"name":"` + big + `"}`)
	var input struct {
		Name string `json:"name" validate:"required"`
	}
	err := bind.Body(req).Limit(10).Bind(&input)
	if err == nil {
		t.Error("expected error due to body size limit")
	}
}
