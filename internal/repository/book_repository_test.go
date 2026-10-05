package repository_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

func newBook(title, category string) *model.Book {
	return &model.Book{
		Title:       title,
		Author:      "Some Author",
		ISBN:        "isbn-" + title,
		Category:    category,
		Description: strPtr("description of " + title),
	}
}

func seed(t *testing.T, repo repository.BookRepository, n int, category string) {
	t.Helper()
	for i := 0; i < n; i++ {
		if err := repo.Create(newBook(fmt.Sprintf("%s-%d", category, i), category)); err != nil {
			t.Fatalf("seed: create failed: %v", err)
		}
	}
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

// ---------- Create ----------

func TestCreate_AssignsSequentialIDsStartingFromOne(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	a, b := newBook("a", "x"), newBook("b", "x")

	if err := repo.Create(a); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(b); err != nil {
		t.Fatal(err)
	}

	if a.ID != 1 || b.ID != 2 {
		t.Errorf("expected IDs 1 and 2, got %d and %d", a.ID, b.ID)
	}
}

func TestCreate_IDsAreNotReusedAfterDelete(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	seed(t, repo, 2, "x")

	if err := repo.Delete(2); err != nil {
		t.Fatal(err)
	}
	c := newBook("c", "x")
	if err := repo.Create(c); err != nil {
		t.Fatal(err)
	}

	if c.ID != 3 {
		t.Errorf("expected ID 3 after deleting ID 2, got %d", c.ID)
	}
}


// Regression test: in-memory Create used to leave them as 0001-01-01.
func TestCreate_SetsCreatedAtAndUpdatedAt(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	b := newBook("timed", "x")

	if err := repo.Create(b); err != nil {
		t.Fatal(err)
	}

	if b.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero after Create")
	}
	if b.UpdatedAt.IsZero() {
		t.Error("UpdatedAt is zero after Create")
	}
}

// ---------- FindByID ----------

func TestFindByID_ReturnsStoredBook(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	b := newBook("found", "x")
	if err := repo.Create(b); err != nil {
		t.Fatal(err)
	}

	got, err := repo.FindByID(b.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "found" || got.ID != b.ID {
		t.Errorf("unexpected book: %+v", got)
	}
}

func TestFindByID_NotFound(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()

	for _, id := range []uint{0, 1, 999} {
		got, err := repo.FindByID(id)
		if !errors.Is(err, repository.ErrNotFound) {
			t.Errorf("id %d: expected ErrNotFound, got %v", id, err)
		}
		if got != nil {
			t.Errorf("id %d: expected nil book, got %+v", id, got)
		}
	}
}

// FindByID must return a copy: changing the result must not modify stored data.
// Regression test for returning pointers to internal objects.
func TestFindByID_ReturnsCopy(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	b := newBook("original", "x")
	if err := repo.Create(b); err != nil {
		t.Fatal(err)
	}
	id := b.ID

	first, _ := repo.FindByID(id)
	first.Title = "hacked"

	second, _ := repo.FindByID(id)
	if second.Title != "original" {
		t.Errorf("stored book was modified through a returned pointer: title = %q", second.Title)
	}
}

// ---------- Update ----------

func TestUpdate_CopiesFieldsKeepsIDAndSetsUpdatedAt(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	orig := newBook("old", "old-cat")
	if err := repo.Create(orig); err != nil {
		t.Fatal(err)
	}
	id := orig.ID
	before := time.Now()

	err := repo.Update(id, &model.Book{
		ID:          99, // must not overwrite the real ID
		Title:       "new title",
		Author:      "new author",
		ISBN:        "new-isbn",
		Category:    "new-cat",
		Description:   strPtr("new description"),
		PublishedYear: intPtr(2020),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != id {
		t.Errorf("ID changed to %d", got.ID)
	}
	if got.Title != "new title" || got.Author != "new author" || got.ISBN != "new-isbn" ||
		got.Category != "new-cat" ||
		got.Description == nil || *got.Description != "new description" ||
		got.PublishedYear == nil || *got.PublishedYear != 2020 {
		t.Errorf("fields were not updated: %+v", got)
	}
	if got.UpdatedAt.Before(before) {
		t.Errorf("UpdatedAt was not refreshed: %v", got.UpdatedAt)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()

	err := repo.Update(42, newBook("x", "x"))

	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ---------- Delete ----------

func TestDelete_RemovesBook(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	b := newBook("gone", "x")
	if err := repo.Create(b); err != nil {
		t.Fatal(err)
	}

	if err := repo.Delete(b.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := repo.FindByID(b.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestDelete_NotFoundAndDoubleDelete(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	b := newBook("once", "x")
	if err := repo.Create(b); err != nil {
		t.Fatal(err)
	}

	if err := repo.Delete(999); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("unknown id: expected ErrNotFound, got %v", err)
	}
	if err := repo.Delete(b.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(b.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("second delete: expected ErrNotFound, got %v", err)
	}
}

// ---------- List ----------

func TestList_EmptyRepository(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()

	got, err := repo.List(nil, 1, 10)

	if err != nil || len(got) != 0 {
		t.Errorf("expected empty list and nil error, got %d items, err=%v", len(got), err)
	}
}

func TestList_NilCategoryReturnsAll(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	seed(t, repo, 3, "a")
	seed(t, repo, 2, "b")

	got, err := repo.List(nil, 1, 10)

	if err != nil || len(got) != 5 {
		t.Errorf("expected 5 books, got %d (err=%v)", len(got), err)
	}
}

func TestList_FiltersByCategory(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	seed(t, repo, 3, "a")
	seed(t, repo, 2, "b")

	got, err := repo.List(strPtr("b"), 1, 10)

	if err != nil || len(got) != 2 {
		t.Fatalf("expected 2 books, got %d (err=%v)", len(got), err)
	}
	for _, b := range got {
		if b.Category != "b" {
			t.Errorf("book from wrong category: %+v", b)
		}
	}
}

func TestList_UnknownCategoryReturnsEmpty(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	seed(t, repo, 3, "a")

	got, err := repo.List(strPtr("nope"), 1, 10)

	if err != nil || len(got) != 0 {
		t.Errorf("expected empty result, got %d (err=%v)", len(got), err)
	}
}

// The handler used to pass &category even when the parameter was missing,
// so an empty string must mean "no filter" instead of Category == "".
// Regression test for that bug.
func TestList_EmptyCategoryMeansNoFilter(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	seed(t, repo, 3, "a")
	seed(t, repo, 2, "b")

	got, err := repo.List(strPtr(""), 1, 10)

	if err != nil || len(got) != 5 {
		t.Errorf("expected 5 books for empty category, got %d (err=%v)", len(got), err)
	}
}

func TestList_Pagination(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	seed(t, repo, 5, "a")

	cases := []struct {
		name        string
		page, limit int
		want        int
	}{
		{"first page", 1, 2, 2},
		{"second page", 2, 2, 2},
		{"last partial page", 3, 2, 1},
		{"page beyond the end", 4, 2, 0},
		{"limit bigger than total", 1, 10, 5},
		{"limit zero", 1, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.List(nil, tc.page, tc.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tc.want {
				t.Errorf("page=%d limit=%d: expected %d books, got %d", tc.page, tc.limit, tc.want, len(got))
			}
		})
	}
}

// page <= 0 or limit < 0 used to produce negative slice bounds and a panic
// (a 500 in production via Recovery). List must validate its input.
func TestList_InvalidPaginationDoesNotPanic(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	seed(t, repo, 3, "a")

	cases := []struct {
		name        string
		page, limit int
	}{
		{"page zero", 0, 10},
		{"negative page", -1, 2},
		{"negative limit", 1, -1},
		{"both negative", -1, -5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("List(page=%d, limit=%d) panicked: %v", tc.page, tc.limit, r)
				}
			}()
			_, _ = repo.List(nil, tc.page, tc.limit)
		})
	}
}

// Map iteration order is random; List must sort by ID.
func TestList_OrderIsDeterministicByID(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	seed(t, repo, 50, "a")

	got, err := repo.List(nil, 1, 50)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].ID >= got[i].ID {
			t.Fatalf("list is not sorted by ID: %d comes before %d", got[i-1].ID, got[i].ID)
		}
	}
}

// Consequence of unstable order: pages overlapped and some books were never shown.
func TestList_PagesCoverAllBooksExactlyOnce(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	seed(t, repo, 30, "a")

	seen := map[uint]int{}
	for page := 1; page <= 3; page++ {
		got, err := repo.List(nil, page, 10)
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range got {
			seen[b.ID]++
		}
	}

	if len(seen) != 30 {
		t.Errorf("pages cover %d distinct books, want 30", len(seen))
	}
	for id, n := range seen {
		if n > 1 {
			t.Errorf("book %d appeared on %d pages", id, n)
		}
	}
}

// ---------- Concurrency ----------

// Run with `go test -race`. Readers must not touch internal objects outside the mutex,
// so the repository returns copies and the race detector stays quiet
// while Update modifies the stored books.
func TestConcurrentAccess_NoDataRace(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	seed(t, repo, 5, "a")

	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				_ = repo.Update(1, newBook(fmt.Sprintf("w%d-%d", w, i), "a"))
				_ = repo.Create(newBook("c", "a"))
				_ = repo.Delete(uint(100 + i))
			}
		}(w)
	}
	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			total := 0
			for i := 0; i < 200; i++ {
				if b, err := repo.FindByID(1); err == nil {
					total += len(b.Title)
				}
				books, _ := repo.List(nil, 1, 10)
				for _, b := range books {
					total += len(b.Title)
				}
			}
			if total < 0 {
				t.Error("unreachable")
			}
		}()
	}
	wg.Wait()
}

func TestCreate_DoesNotAliasCallerBook(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	b := newBook("original", "x")
	if err := repo.Create(b); err != nil {
		t.Fatal(err)
	}

	b.Title = "changed by caller"
	*b.Description = "changed by caller"

	got, _ := repo.FindByID(b.ID)
	if got.Title != "original" {
		t.Errorf("stored title changed through caller's pointer: %q", got.Title)
	}
	if *got.Description != "description of original" {
		t.Errorf("stored description changed through caller's pointer: %q", *got.Description)
	}
}

func TestUpdate_DoesNotAliasRequestPointers(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	b := newBook("t", "x")
	if err := repo.Create(b); err != nil {
		t.Fatal(err)
	}
	desc := "from request"
	req := newBook("t2", "x")
	req.Description = &desc
	if err := repo.Update(b.ID, req); err != nil {
		t.Fatal(err)
	}

	desc = "mutated after update"

	got, _ := repo.FindByID(b.ID)
	if *got.Description != "from request" {
		t.Errorf("stored description shares memory with request: %q", *got.Description)
	}
}

func TestList_HugePageAndLimitDoNotOverflow(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	seed(t, repo, 3, "a")
	const maxInt = int(^uint(0) >> 1)

	for _, tc := range []struct{ page, limit int }{{maxInt, 10}, {2, maxInt}, {maxInt, maxInt}} {
		got, err := repo.List(nil, tc.page, tc.limit)
		if err != nil {
			t.Fatalf("page=%d limit=%d: %v", tc.page, tc.limit, err)
		}
		if tc.page == 2 && len(got) != 0 {
			t.Errorf("page=2 limit=maxInt: expected 0 books, got %d", len(got))
		}
	}
}