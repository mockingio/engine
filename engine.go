package engine

import (
	"context"
	"github.com/pkg/errors"
	"net/http"
	"time"

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

func (eng *Engine) Handler(w http.ResponseWriter, r *http.Request) {
	if eng.isPaused {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	response := eng.Match(r)
	if response == nil {
		// TODO: no matched? What will be the response?
		w.WriteHeader(http.StatusNotFound)
		return
	}

	for k, v := range response.Headers {
		w.Header().Add(k, v)
	}

	w.WriteHeader(response.Status)
	_, _ = w.Write([]byte(response.Body))
}
