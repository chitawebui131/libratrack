package model_test

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/gin-gonic/gin/binding"

	"libratrack/internal/model"
)

func validCreate() model.CreateBookRequest {
	return model.CreateBookRequest{Title: "T", ISBN: "I", Author: "A", Category: "C"}
}

func validUpdate() model.UpdateBookRequest {
	return model.UpdateBookRequest{Title: "T", ISBN: "I", Author: "A", Category: "C"}
}

// ---------- CreateBookRequest ----------

func TestCreateBookRequest_ValidWithOnlyRequiredFields(t *testing.T) {
	req := validCreate()
	if err := binding.Validator.ValidateStruct(&req); err != nil {
		t.Errorf("expected valid request, got %v", err)
	}
}

func TestCreateBookRequest_ValidWithOptionalFields(t *testing.T) {
	year, desc := 2015, "text"
	req := validCreate()
	req.PublishedYear, req.Description = &year, &desc

	if err := binding.Validator.ValidateStruct(&req); err != nil {
		t.Errorf("expected valid request, got %v", err)
	}
}

func TestCreateBookRequest_EmptyIsInvalid(t *testing.T) {
	if err := binding.Validator.ValidateStruct(&model.CreateBookRequest{}); err == nil {
		t.Error("expected validation error for empty request")
	}
}

func TestCreateBookRequest_EachRequiredFieldIsEnforced(t *testing.T) {
	cases := map[string]func(*model.CreateBookRequest){
		"title":    func(r *model.CreateBookRequest) { r.Title = "" },
		"isbn":     func(r *model.CreateBookRequest) { r.ISBN = "" },
		"author":   func(r *model.CreateBookRequest) { r.Author = "" },
		"category": func(r *model.CreateBookRequest) { r.Category = "" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			req := validCreate()
			mutate(&req)
			if err := binding.Validator.ValidateStruct(&req); err == nil {
				t.Errorf("expected validation error when %s is empty", name)
			}
		})
	}
}

// binding:"required" alone accepts whitespace-only strings;
// the nonblank validator must reject them.
func TestCreateBookRequest_RejectsBlankValues(t *testing.T) {
	cases := map[string]func(*model.CreateBookRequest){
		"title":    func(r *model.CreateBookRequest) { r.Title = "   " },
		"isbn":     func(r *model.CreateBookRequest) { r.ISBN = "   " },
		"author":   func(r *model.CreateBookRequest) { r.Author = "   " },
		"category": func(r *model.CreateBookRequest) { r.Category = "   " },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			req := validCreate()
			mutate(&req)
			if err := binding.Validator.ValidateStruct(&req); err == nil {
				t.Errorf("whitespace-only %s was accepted", name)
			}
		})
	}
}

// ---------- UpdateBookRequest ----------

func TestUpdateBookRequest_ValidWithOnlyRequiredFields(t *testing.T) {
	req := validUpdate()
	if err := binding.Validator.ValidateStruct(&req); err != nil {
		t.Errorf("expected valid request, got %v", err)
	}
}

func TestUpdateBookRequest_EachRequiredFieldIsEnforced(t *testing.T) {
	cases := map[string]func(*model.UpdateBookRequest){
		"title":    func(r *model.UpdateBookRequest) { r.Title = "" },
		"isbn":     func(r *model.UpdateBookRequest) { r.ISBN = "" },
		"author":   func(r *model.UpdateBookRequest) { r.Author = "" },
		"category": func(r *model.UpdateBookRequest) { r.Category = "" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			req := validUpdate()
			mutate(&req)
			if err := binding.Validator.ValidateStruct(&req); err == nil {
				t.Errorf("expected validation error when %s is empty", name)
			}
		})
	}
}

// ---------- JSON shape of Book ----------

func TestCreateBookRequest_DecodesSnakeCaseJSON(t *testing.T) {
	var req model.CreateBookRequest
	body := `{"title":"T","isbn":"I","author":"A","category":"C","published_year":2015,"description":"D"}`

	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatal(err)
	}

	if req.Title != "T" || req.ISBN != "I" || req.Author != "A" || req.Category != "C" {
		t.Errorf("required fields not decoded: %+v", req)
	}
	if req.PublishedYear == nil || *req.PublishedYear != 2015 {
		t.Errorf("published_year not decoded: %v", req.PublishedYear)
	}
	if req.Description == nil || *req.Description != "D" {
		t.Errorf("description not decoded: %v", req.Description)
	}
}

// Book must use snake_case JSON keys, consistent with the request structs.
// Without json tags the keys would be "ID", "CreatedAt", etc.
func TestBook_JSONKeysAreSnakeCase(t *testing.T) {
	data, err := json.Marshal(model.Book{})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	want := []string{"id", "title", "isbn", "author", "category", "published_year", "description", "created_at", "updated_at"}
	for _, key := range want {
		if _, ok := got[key]; !ok {
			keys := make([]string, 0, len(got))
			for k := range got {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			t.Errorf("missing JSON key %q; actual keys: %v", key, keys)
			return
		}
	}
}