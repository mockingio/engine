package smocky

import (
	"context"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mockingio/engine/engine"
	"github.com/mockingio/engine/engine/mock"
	"github.com/mockingio/engine/engine/persistent/memory"
)

const (
	Header        = "header"
	Body          = "body"
	QueryString   = "query_string"
	Cookie        = "cookie"
	RouteParam    = "route_param"
	RequestNumber = "request_number"
)

const (
	Equal = "equal"
	Regex = "regex"
)

type Headers map[string]string

func New() *Builder {
	return &Builder{
		config: &mock.Mock{},
	}
}

type Builder struct {
	response *mock.Response
	route    *mock.Route
	config   *mock.Mock
}

func (b *Builder) Start(t *testing.T) *httptest.Server {
	b.clear()
	if err := b.config.Validate(); err != nil {
		t.Errorf("invalid config: %v", err)
	}
	id := uuid.NewString()
	b.config.ID = id

	mem := memory.New()
	_ = mem.SetMock(context.Background(), b.config)
	_ = mem.SetActiveSession(context.Background(), id, "session-id")

	m := engine.New(id, mem)

	return httptest.NewServer(http.HandlerFunc(m.Handler))
}

func (b *Builder) Port(port string) *Builder {
	b.config.Port = port
	return b
}

func (b *Builder) Post(url string) *Builder {
	b.clear()
	b.route = &mock.Route{
		Method: "POST",
		Path:   url,
	}
	return b
}

func (b *Builder) Get(url string) *Builder {
	b.clear()
	b.route = &mock.Route{
		Method: "GET",
		Path:   url,
	}
	return b
}

func (b *Builder) Put(url string) *Builder {
	b.clear()
	b.route = &mock.Route{
		Method: "PUT",
		Path:   url,
	}
	return b
}

func (b *Builder) Delete(url string) *Builder {
	b.clear()
	b.route = &mock.Route{
		Method: "DELETE",
		Path:   url,
	}
	return b
}

func (b *Builder) Option(url string) *Builder {
	b.clear()
	b.route = &mock.Route{
		Method: "OPTION",
		Path:   url,
	}
	return b
}

func (b *Builder) Response(status int, body string, headers ...Headers) *Response {
	if len(headers) == 0 {
		headers = append(headers, map[string]string{})
	}

	b.response = &mock.Response{
		Body:    body,
		Status:  status,
		Headers: headers[0],
	}

	resp := &Response{
		builder: b,
	}

	return resp
}

func (b *Builder) clear() {
	if b.response != nil {
		b.route.Responses = append(b.route.Responses, *b.response)
		b.response = nil
	}

	if b.route != nil {
		b.config.Routes = append(b.config.Routes, b.route)
		b.route = nil
	}
}
