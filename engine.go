package engine

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/mockingio/engine/matcher"
	"github.com/mockingio/engine/mock"
	"github.com/mockingio/engine/persistent"
)

type Engine struct {
	mockID   string
	isPaused bool
	db       persistent.Persistent
	mock     *mock.Mock
}

func New(mockID string, db persistent.Persistent) *Engine {
	return &Engine{
		mockID: mockID,
		db:     db,
	}
}

func (eng *Engine) Resume() {
	eng.isPaused = false
}

func (eng *Engine) Pause() {
	eng.isPaused = true
}

func (eng *Engine) Match(req *http.Request) *mock.Response {
	ctx := req.Context()
	if err := eng.reloadMock(ctx); err != nil {
		log.WithError(err).Error("reload mock")
		return nil
	}

	mok := eng.getMock()

	sessionID, err := eng.db.GetActiveSession(ctx, eng.mockID)
	if err != nil {
		log.WithError(err).WithField("config_id", eng.mockID).Error("get active session")
	}

	for _, route := range mok.Routes {
		log.Debugf("Matching route: %v %v", route.Method, route.Path)
		response, err := matcher.NewRouteMatcher(route, matcher.Context{
			HTTPRequest: req,
			SessionID:   sessionID,
		}, eng.db).Match()
		if err != nil {
			log.WithError(err).Error("matching route")
			continue
		}

		if response == nil {
			log.Debug("no route matched")
			continue
		}

		if response.Delay > 0 {
			time.Sleep(time.Millisecond * time.Duration(response.Delay))
		}

		return response
	}

	return nil
}

func (eng *Engine) Handler(w http.ResponseWriter, r *http.Request) {
	if eng.isPaused {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	response := eng.Match(r)
	if response == nil {
		mok := eng.getMock()

		if mok.AutoCORS && r.Method == http.MethodOptions {
			eng.corsHandler(w, r)
			return
		}

		if mok.ProxyEnabled() {
			eng.proxyHandler(w, r)
			return
		}

		eng.noMatchHandler(w)
		return
	}

	for k, v := range response.Headers {
		w.Header().Add(k, v)
	}

	w.WriteHeader(response.Status)
	_, _ = w.Write([]byte(response.Body))
}

func (eng *Engine) noMatchHandler(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func (eng *Engine) corsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (eng *Engine) proxyHandler(w http.ResponseWriter, r *http.Request) {
	proxy := eng.getMock().Proxy

	req, err := copyProxyRequest(r, proxy)
	if err != nil {
		log.WithError(err).Error("copy request")
		eng.noMatchHandler(w)
		return
	}

	res, err := (&http.Client{}).Do(req)
	if err != nil {
		log.WithError(err).Error("make proxy request")
		eng.noMatchHandler(w)
		return
	}
	defer func() { _ = res.Body.Close() }()

	writeProxyResponse(res, w, proxy)
}

func (eng *Engine) getMock() *mock.Mock {
	return eng.mock
}

func (eng *Engine) reloadMock(ctx context.Context) error {
	mok, err := eng.db.GetMock(ctx, eng.mockID)
	if err != nil {
		return errors.Wrap(err, "get mock from DB")
	}

	if mok == nil {
		return errors.New("mock not found")
	}

	eng.mock = mok

	return nil
}

func copyProxyRequest(r *http.Request, proxy *mock.Proxy) (*http.Request, error) {
	req, err := http.NewRequest(r.Method, proxy.Host+r.URL.Path, r.Body)
	if err != nil {
		return nil, err
	}
	req.Header = r.Header

	for k, v := range proxy.RequestHeaders {
		req.Header.Add(k, v)
	}

	return req, nil
}

func writeProxyResponse(res *http.Response, w http.ResponseWriter, proxy *mock.Proxy) {
	for k, v := range res.Header {
		w.Header().Add(k, v[0])
	}
	for k, v := range proxy.ResponseHeaders {
		w.Header().Add(k, v)
	}

	w.WriteHeader(res.StatusCode)
	_, _ = io.Copy(w, res.Body)
}
