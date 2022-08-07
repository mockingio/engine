package mock_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/mockingio/engine/mock"
)

func TestRoute_Clone(t *testing.T) {
	route := Route{
		ID:        "",
		Method:    "POST",
		Path:      "/",
		Responses: []Response{{Status: http.StatusOK}},
	}

	clone := route.Clone()
	assert.True(t, clone.Validate() == nil)
	assert.NotEqual(t, route.ID, clone.ID)

	// TODO: Should compare pointer for all property
}

func TestRoute_Validate(t *testing.T) {
	validResponse := []Response{{Status: http.StatusOK}}

	tests := []struct {
		name  string
		route Route
		error bool
	}{
		{"valid route", Route{Method: "POST", Path: "/", Responses: validResponse}, false},
		{"invalid route, missing request", Route{Responses: validResponse}, true},
		{"invalid route, missing response", Route{Method: "POST", Path: "/"}, true},
		{"invalid route, invalid response", Route{Method: "POST", Path: "/", Responses: []Response{}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.route.Validate()
			assert.Equal(t, tt.error, err != nil)
		})
	}
}
