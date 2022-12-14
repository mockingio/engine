package memory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mockingio/engine/mock"
	. "github.com/mockingio/engine/persistent/memory"
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

	value, err = m.GetMock(context.Background(), "*random*")
	require.NoError(t, err)
	assert.Nil(t, value)
}

func TestMemory_GetInt(t *testing.T) {
	m := New()

	err := m.Set(context.Background(), "*id*", 200)
	require.NoError(t, err)

	value, err := m.GetInt(context.Background(), "*id*")
	require.NoError(t, err)
	assert.Equal(t, value, 200)

	value, err = m.GetInt(context.Background(), "*random*")
	require.NoError(t, err)
	assert.Equal(t, value, 0)
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

	// when key does not exist
	val, err = m.Increment(context.Background(), "*random*")
	require.NoError(t, err)
	assert.Equal(t, 1, val)

	// when value is not int
	err = m.Set(context.Background(), "*non-int*", "200")
	require.NoError(t, err)
	_, err = m.Increment(context.Background(), "*non-int*")
	assert.Error(t, err)
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

	t.Run("success", func(t *testing.T) {
		err := m.PatchRoute(context.Background(), "mockid", "routeid", `{"method": "POST"}`)
		require.NoError(t, err)
		assert.Equal(t, "POST", mok.Routes[0].Method)
	})

	t.Run("mock not found", func(t *testing.T) {
		err := m.PatchRoute(context.Background(), "random", "", `{}`)
		require.Error(t, err)
	})

	t.Run("route not found", func(t *testing.T) {
		err := m.PatchRoute(context.Background(), "mockid", "random", `{}`)
		require.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		err := m.PatchRoute(context.Background(), "mockid", "routeid", `{"method": "}`)
		require.Error(t, err)
	})
}

func TestMemory_DeleteRoute(t *testing.T) {
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

	t.Run("success", func(t *testing.T) {
		err := m.DeleteRoute(context.Background(), "mockid", "routeid")
		require.NoError(t, err)

		configs, err := m.GetMock(context.Background(), "mockid")
		require.NoError(t, err)
		assert.Equal(t, 1, len(configs.Routes))
	})

	t.Run("mock not found", func(t *testing.T) {
		err := m.DeleteRoute(context.Background(), "random", "")
		assert.Error(t, err)
	})

	t.Run("route not found", func(t *testing.T) {
		err := m.DeleteRoute(context.Background(), "mockid", "random")
		assert.Error(t, err)
	})
}

func TestMemory_CreateRoute(t *testing.T) {
	m := New()
	mok := &mock.Mock{
		ID: "mockid",
		Routes: []*mock.Route{
			{
				ID:     "routeid",
				Method: "GET",
			},
		},
	}
	_ = m.SetMock(context.Background(), mok)

	t.Run("success", func(t *testing.T) {
		err := m.CreateRoute(context.Background(), "mockid", `{"id":"routeid1","method":"PUT"}`)
		require.NoError(t, err)
		configs, err := m.GetMock(context.Background(), "mockid")
		require.NoError(t, err)
		assert.Equal(t, 2, len(configs.Routes))
	})

	t.Run("mock not found", func(t *testing.T) {
		err := m.CreateRoute(context.Background(), "random", `{}`)
		assert.Error(t, err)
	})

	t.Run("route already created", func(t *testing.T) {
		err := m.CreateRoute(context.Background(), "mockid", `{"id":"routeid1","method":"PUT"}`)
		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		err := m.CreateRoute(context.Background(), "mockid", `{"method": "}`)
		assert.Error(t, err)
	})

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

	t.Run("success", func(t *testing.T) {
		err := m.PatchResponse(context.Background(), "mockid", "routeid2", "responseid2", `{"status": 201}`)
		require.NoError(t, err)
		assert.Equal(t, 201, mok.Routes[1].Responses[1].Status)
	})

	t.Run("mock not found", func(t *testing.T) {
		err := m.PatchResponse(context.Background(), "random", "", "", `{}`)
		assert.Error(t, err)
	})

	t.Run("route not found", func(t *testing.T) {
		err := m.PatchResponse(context.Background(), "mockid", "random", "responseid2", `{}`)
		assert.Error(t, err)
	})

	t.Run("response not found", func(t *testing.T) {
		err := m.PatchResponse(context.Background(), "mockid", "routeid2", "random", `{`)
		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		err := m.PatchResponse(context.Background(), "mockid", "routeid2", "responseid2", `{": 201}`)
		assert.Error(t, err)
	})
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
	m.SubscribeMockChanges(func(mo mock.Mock) {
		updatedMock = mo
	})
	_ = m.SetMock(context.Background(), cfg)
	assert.Equal(t, updatedMock, *cfg)
}
