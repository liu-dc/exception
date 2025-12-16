package try

import (
	"testing"

	"github.com/liu-dc/exception/throw"
)

func TestTryIndex(t *testing.T) {
	expectedIndex := 123
	caught := false
	var caughtError throw.Error

	Run(func() {
		throw.Index(expectedIndex, "test error")
	}).Index(expectedIndex, func(err throw.Error) {
		caught = true
		caughtError = err
	}).Do()

	if !caught {
		t.Error("Expected Index handler to be called")
	}
	if caughtError.Index != expectedIndex {
		t.Errorf("Expected index %d, got %d", expectedIndex, caughtError.Index)
	}
}

func TestTryGroup(t *testing.T) {
	group := Group{400, 401, 403, 404}
	groupCaught := false
	globalCaught := false

	Run(func() {
		throw.Index(404, "not found")
	}).Group(group, func(err throw.Error) {
		groupCaught = true
		if err.Index != 404 {
			t.Errorf("Expected index 404, got %d", err.Index)
		}
	}).Catch(func(err throw.Error) {
		globalCaught = true
	}).Do()

	if !groupCaught {
		t.Error("Expected Group handler to be called")
	}
	if !globalCaught {
		t.Error("Expected Catch handler to be called")
	}
}

func TestTryFilter(t *testing.T) {
	filterCalled := false
	handlerCalled := false

	Run(func() {
		throw.Index(500, "server error")
	}).Filter(func(err throw.Error) bool {
		filterCalled = true
		// Only allow 4xx errors
		return err.Index >= 400 && err.Index < 500
	}).Index(500, func(err throw.Error) {
		handlerCalled = true
	}).Do()

	if !filterCalled {
		t.Error("Expected Filter to be called")
	}
	if handlerCalled {
		t.Error("Expected handler NOT to be called due to filter")
	}
}

func TestTryMultipleFilters(t *testing.T) {
	filter1Called := false
	filter2Called := false
	handlerCalled := false

	Run(func() {
		throw.IndexWithContext(403, "forbidden", "userID", 123)
	}).Filter(func(err throw.Error) bool {
		filter1Called = true
		return err.Index >= 400
	}).Filter(func(err throw.Error) bool {
		filter2Called = true
		_, exists := err.GetContext("userID")
		return exists
	}).Index(403, func(err throw.Error) {
		handlerCalled = true
	}).Do()

	if !filter1Called {
		t.Error("Expected filter1 to be called")
	}
	if !filter2Called {
		t.Error("Expected filter2 to be called")
	}
	if !handlerCalled {
		t.Error("Expected handler to be called after both filters")
	}
}

func TestTryUnknown(t *testing.T) {
	unknownCaught := false
	globalCaught := false

	Run(func() {
		throw.Index(999, "unknown error")
	}).Index(404, func(err throw.Error) {
		// This should not be called
		t.Error("Unexpected Index handler called")
	}).Unknown(func(err throw.Error) {
		unknownCaught = true
	}).Catch(func(err throw.Error) {
		globalCaught = true
	}).Do()

	if !unknownCaught {
		t.Error("Expected Unknown handler to be called")
	}
	if !globalCaught {
		t.Error("Expected Catch handler to be called")
	}
}

func TestTryFinally(t *testing.T) {
	finallyCalled := false
	handlerCalled := false

	Run(func() {
		throw.Index(500, "test")
	}).Index(500, func(err throw.Error) {
		handlerCalled = true
	}).Finally(func() {
		finallyCalled = true
	})

	if !handlerCalled {
		t.Error("Expected Index handler to be called")
	}
	if !finallyCalled {
		t.Error("Expected Finally to be called")
	}
}

func TestTryDo(t *testing.T) {
	called := false

	Run(func() {
		called = true
	}).Do()

	if !called {
		t.Error("Expected Do to execute the function")
	}
}

func TestTryContextSupport(t *testing.T) {
	contextHandled := false

	Run(func() {
		throw.IndexWithContext(404, "not found", "resource", "user")
	}).Index(404, func(err throw.Error) {
		if resource, exists := err.GetContext("resource"); exists && resource == "user" {
			contextHandled = true
		}
	}).Do()

	if !contextHandled {
		t.Error("Expected context to be passed correctly")
	}
}
