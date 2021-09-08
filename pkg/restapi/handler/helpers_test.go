package handler_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "gitlab.com/umitop/umid/pkg/restapi/handler"
)

func TestParseParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name       string
		TotalCount int
		Offset     int
		Limit      int
		FirstIndex int
		LastIndex  int
		Error      error
	}{
		{"a", 10, 0, 10, 0, 10, nil},
		{"a", 10, 0, 20, 0, 10, nil},
		{"b", 10, -10, 10, 0, 10, nil},
		{"b2", 10, -10, 20, 0, 10, nil},
		{"c", 10, -15, 10, 0, 5, nil},
		{"d", 10, -20, 10, 0, 0, ErrOutOfRange},
		{"e", 10, 15, 10, 0, 0, ErrOutOfRange},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			target := fmt.Sprintf("/?offset=%d&limit=%d", tc.Offset, tc.Limit)
			r := httptest.NewRequest(http.MethodGet, target, nil)
			firstIndex, lastIndex, err := ParseParams(r, 10)

			if firstIndex != tc.FirstIndex {
				t.Errorf("firstIndex expecting %d, got %d", tc.FirstIndex, firstIndex)
			}

			if lastIndex != tc.LastIndex {
				t.Errorf("lastIndex expecting %d, got %d", tc.LastIndex, lastIndex)
			}

			if !errors.Is(err, tc.Error) {
				t.Errorf("err expecting %v, got %v", tc.Error, err)
			}
		})
	}
}
