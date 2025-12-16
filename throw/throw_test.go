package throw

import (
	"errors"
	"testing"
)

func TestIndexWithContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(Error); ok {
				if e.Index != 404 {
					t.Errorf("Expected index 404, got %d", e.Index)
				}
				if e.Message != "not found" {
					t.Errorf("Expected message 'not found', got %s", e.Message)
				}
				if userID, exists := e.GetContext("userID"); !exists || userID != 123 {
					t.Errorf("Expected context userID=123, got %v (exists: %v)", userID, exists)
				}
				if e.Stack == "" {
					t.Error("Expected stack trace, got empty")
				}
			} else {
				t.Errorf("Expected Error type, got %T", r)
			}
		}
	}()

	IndexWithContext(404, "not found", "userID", 123)
}

func TestWithContextMethod(t *testing.T) {
	e := Error{
		Index:   500,
		Message: "server error",
		Stack:   "stack trace",
	}

	// Test single context
	e = e.WithContext("key1", "value1")
	if value, exists := e.GetContext("key1"); !exists || value != "value1" {
		t.Errorf("Expected key1=value1, got %v (exists: %v)", value, exists)
	}

	// Test multiple contexts
	e = e.WithContext("key2", 2)
	if value, exists := e.GetContext("key2"); !exists || value != 2 {
		t.Errorf("Expected key2=2, got %v (exists: %v)", value, exists)
	}

	// Test HasContext
	if !e.HasContext("key1") {
		t.Error("Expected HasContext('key1') to return true")
	}
	if e.HasContext("nonexistent") {
		t.Error("Expected HasContext('nonexistent') to return false")
	}
}

func TestWithContextMap(t *testing.T) {
	e := Error{
		Index:   400,
		Message: "bad request",
	}

	ctx := map[string]interface{}{
		"param1": "value1",
		"param2": 2,
	}

	e = e.WithContextMap(ctx)

	for k, v := range ctx {
		if value, exists := e.GetContext(k); !exists || value != v {
			t.Errorf("Expected %s=%v, got %v (exists: %v)", k, v, value, exists)
		}
	}
}

func TestErrWithContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(Error); ok {
				if e.Message != "test error" {
					t.Errorf("Expected message 'test error', got %s", e.Message)
				}
				if requestID, exists := e.GetContext("requestID"); !exists || requestID != "req-123" {
					t.Errorf("Expected context requestID=req-123, got %v (exists: %v)", requestID, exists)
				}
			}
		}
	}()

	err := errors.New("test error")
	ErrWithContext(err, "requestID", "req-123")
}

func TestTrueWithContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(Error); ok {
				if e.Message != "should panic" {
					t.Errorf("Expected message 'should panic', got %s", e.Message)
				}
				if condition, exists := e.GetContext("condition"); !exists || condition != true {
					t.Errorf("Expected context condition=true, got %v (exists: %v)", condition, exists)
				}
			}
		}
	}()

	TrueWithContext(true, "should panic", "condition", true)
}
