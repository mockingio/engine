package memory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuongaz/smocky-engine/engine/mock"
	. "github.com/tuongaz/smocky-engine/engine/persistent/memory"
)

func TestMemory_GetSetConfig(t *testing.T) {
	cfg := &mock.Mock{
		Port: "1234",
		ID:   "*id*",
	}

	m := New()

	err := m.SetMock(context.Background(), cfg)
	require.NoError(t, err)

	value, err := m.GetMock(context.Background(), "*id*")

	require.NoError(t, err)
	assert.Equal(t, value, cfg)
}

func TestMemory_GetInt(t *testing.T) {
	m := New()

	err := m.Set(context.Background(), "*id*", 200)
	require.NoError(t, err)

	value, err := m.GetInt(context.Background(), "*id*")

	require.NoError(t, err)
	assert.Equal(t, value, 200)
}

func TestMemory_Increase(t *testing.T) {
	m := New()

	err := m.Set(context.Background(), "*id*", 200)
	require.NoError(t, err)

	val, err := m.Increment(context.Background(), "*id*")
	require.NoError(t, err)
	assert.Equal(t, 201, val)

	i, err := m.GetInt(context.Background(), "*id*")
	require.NoError(t, err)
	assert.Equal(t, 201, i)
}

func TestMemory_SetGetActiveSession(t *testing.T) {
	m := New()

	err := m.SetActiveSession(context.Background(), "mockid", "123456")
	require.NoError(t, err)

	v, err := m.GetActiveSession(context.Background(), "mockid")
	require.NoError(t, err)
	assert.Equal(t, "123456", v)
}

func TestMemory_PatchRoute(t *testing.T) {
	m := New()
	mok := &mock.Mock{
		ID: "mockid",
		Routes: []*mock.Route{
			{
				ID:     "routeid",
				Method: "GET",
			},
			{
				ID:     "routeid1",
				Method: "PUT",
			},
		},
	}
	_ = m.SetMock(context.Background(), mok)

	err := m.PatchRoute(context.Background(), "mockid", "routeid", `{"method": "POST"}`)
	require.NoError(t, err)
	assert.Equal(t, "POST", mok.Routes[0].Method)
}

func TestMemory_PatchResponse(t *testing.T) {
	m := New()
	mok := &mock.Mock{
		ID: "mockid",
		Routes: []*mock.Route{
			{
				ID:     "routeid1",
				Method: "GET",
			},
			{
				ID:     "routeid2",
				Method: "PUT",
				Responses: []mock.Response{
					{
						ID:     "responseid1",
						Status: 200,
					},
					{
						ID:     "responseid2",
						Status: 400,
					},
				},
			},
		},
	}
	_ = m.SetMock(context.Background(), mok)

	err := m.PatchResponse(context.Background(), "mockid", "routeid2", "responseid2", `{"status": 201}`)
	require.NoError(t, err)
	assert.Equal(t, 201, mok.Routes[1].Responses[1].Status)
}

func TestMemory_GetConfigs(t *testing.T) {
	cfg1 := &mock.Mock{
		Port: "1234",
		ID:   "*id1*",
	}

	cfg2 := &mock.Mock{
		Port: "1234",
		ID:   "*id2*",
	}

	m := New()
	_ = m.SetMock(context.Background(), cfg1)
	_ = m.SetMock(context.Background(), cfg2)

	configs, err := m.GetMocks(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, len(configs))
}

func TestMemory_OnMockChanges(t *testing.T) {
	cfg := &mock.Mock{
		Port: "1234",
		ID:   "*id1*",
	}
	updatedMock := mock.Mock{}

	m := New()
	m.OnMockChanges(func(mo mock.Mock) {
		updatedMock = mo
	})
	_ = m.SetMock(context.Background(), cfg)
	assert.Equal(t, updatedMock, *cfg)
}
