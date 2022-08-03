package engine_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mockingio/engine"
	"github.com/mockingio/engine/mock"
	"github.com/mockingio/engine/persistent"
	"github.com/mockingio/engine/persistent/memory"
)

func TestEngine_Pause(t *testing.T) {
	eng := engine.New("mock-id", memory.New())
	eng.Pause()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	eng.Handler(w, req)
	res := w.Result()
	assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
}

func TestEngine_PauseResume(t *testing.T) {
	mem := setupMock()
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	eng := engine.New("mock-id", mem)

	eng.Pause()
	w := httptest.NewRecorder()
	eng.Handler(w, req)
	res := w.Result()
	assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)

	eng.Resume()
	w = httptest.NewRecorder()
	eng.Handler(w, req)
	res = w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestEngine_Match(t *testing.T) {
	mem := setupMock()
	eng := engine.New("mock-id", mem)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	eng.Handler(w, req)
	res := w.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	bod, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	assert.Equal(t, "Hello World", string(bod))
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
	assert.Equal(t, "test", res.Header.Get("X-Test"))
}

func TestEngine_Match_With_Delay_Response(t *testing.T) {
	mem := memory.New()
	_ = mem.SetMock(context.Background(), &mock.Mock{
		ID: "mock-id",
		Routes: []*mock.Route{
			{
				Method: "GET",
				Path:   "/hello",
				Responses: []mock.Response{
					{
						Status: 200,
						Delay:  50,
					},
				},
			},
		},
	})

	eng := engine.New("mock-id", mem)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()

	timer := time.Now()
	eng.Handler(w, req)
	res := w.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	assert.True(t, time.Since(timer) > 50*time.Millisecond)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestEngine_NoResponses(t *testing.T) {
	mem := memory.New()
	_ = mem.SetMock(context.Background(), &mock.Mock{
		ID: "mock-id",
		Routes: []*mock.Route{
			{
				Method:    "GET",
				Path:      "/hello",
				Responses: []mock.Response{},
			},
		},
	})

	eng := engine.New("mock-id", mem)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	eng.Handler(w, req)
	res := w.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func setupMock() persistent.Persistent {
	mok := &mock.Mock{
		ID: "mock-id",
		Routes: []*mock.Route{
			{
				Method: "GET",
				Path:   "/hello",
				Responses: []mock.Response{
					{
						Status: 200,
						Body:   "Hello World",
						Headers: map[string]string{
							"Content-Type": "text/plain",
							"X-Test":       "test",
						},
					},
				},
			},
		},
	}
	mem := memory.New()
	_ = mem.SetMock(context.Background(), mok)

	return mem
}
