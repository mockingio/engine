package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/echlebek/patch"
	"github.com/samber/lo"
	"github.com/tuongaz/smocky-engine/engine/mock"
	"github.com/tuongaz/smocky-engine/engine/persistent"
)

var _ persistent.Persistent = &Memory{}

type Memory struct {
	mu      sync.Mutex
	configs map[string]*mock.Mock
	kv      map[string]any
}

func New() *Memory {
	return &Memory{
		configs: map[string]*mock.Mock{},
		kv:      map[string]any{},
	}
}

func (m *Memory) Get(_ context.Context, key string) (any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.kv[key], nil
}

func (m *Memory) Set(_ context.Context, key string, value any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.kv[key] = value
	return nil
}

func (m *Memory) SetMock(ctx context.Context, cfg *mock.Mock) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.configs[cfg.ID] = cfg
	return nil
}

func (m *Memory) GetMock(ctx context.Context, id string) (*mock.Mock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg, ok := m.configs[id]
	if !ok {
		return nil, nil
	}

	return cfg, nil
}

func (m *Memory) GetMocks(ctx context.Context) ([]*mock.Mock, error) {
	var configs []*mock.Mock
	for _, cfg := range m.configs {
		configs = append(configs, cfg)
	}

	return configs, nil
}

func (m *Memory) GetInt(ctx context.Context, key string) (int, error) {
	v, err := m.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	if v == nil {
		return 0, nil
	}

	value, ok := v.(int)
	if !ok {
		return 0, nil
	}

	return value, nil
}

func (m *Memory) Increment(_ context.Context, key string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	value, ok := m.kv[key]
	if !ok {
		m.kv[key] = 1
		return 1, nil
	}

	val, ok := value.(int)
	if !ok {
		return 0, errors.New(fmt.Sprintf("unable to increase non-int key (%s)", key))
	}

	val++
	m.kv[key] = val

	return val, nil
}

func (m *Memory) SetActiveSession(ctx context.Context, mockID string, sessionID string) error {
	return m.Set(ctx, toActiveSessionKey(mockID), sessionID)
}

func (m *Memory) GetActiveSession(ctx context.Context, mockID string) (string, error) {
	value, err := m.Get(ctx, toActiveSessionKey(mockID))
	if err != nil {
		return "", err
	}

	if v, ok := value.(string); ok {
		return v, nil
	}

	return "", errors.New("unable to convert to string value")
}

func (m *Memory) PatchRoute(ctx context.Context, mockID string, routeID string, data string) error {
	mok, err := m.GetMock(ctx, mockID)
	if err != nil {
		return err
	}

	if mok == nil {
		return errors.New("mock not found")
	}

	route, idx, ok := lo.FindIndexOf[*mock.Route](mok.Routes, func(route *mock.Route) bool {
		return route.ID == routeID
	})

	if !ok {
		return errors.New("route not found")
	}

	var values map[string]*json.RawMessage
	if err := json.Unmarshal([]byte(data), &values); err != nil {
		return err
	}

	if err := patch.Struct(route, values); err != nil {
		return err
	}

	mok.Routes[idx] = route

	if err := m.SetMock(ctx, mok); err != nil {
		return err
	}

	return nil
}

func (m *Memory) PatchResponse(ctx context.Context, mockID, routeID, responseID, data string) error {
	mok, err := m.GetMock(ctx, mockID)
	if err != nil {
		return err
	}
	if mok == nil {
		return errors.New("mock not found")
	}

	route, routeIdx, ok := lo.FindIndexOf[*mock.Route](mok.Routes, func(route *mock.Route) bool {
		return route.ID == routeID
	})
	if !ok {
		return errors.New("route not found")
	}

	response, resIdx, ok := lo.FindIndexOf[mock.Response](route.Responses, func(response mock.Response) bool {
		return response.ID == responseID
	})
	if !ok {
		return errors.New("response not found")
	}

	var values map[string]*json.RawMessage
	if err := json.Unmarshal([]byte(data), &values); err != nil {
		return err
	}
	if err := patch.Struct(&response, values); err != nil {
		return err
	}

	route.Responses[resIdx] = response
	mok.Routes[routeIdx] = route

	if err := m.SetMock(ctx, mok); err != nil {
		return err
	}

	return nil
}

func toActiveSessionKey(mockID string) string {
	return fmt.Sprintf("%s-active-session", mockID)
}
