package mock_test

import (
	_ "embed"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"testing"

	. "github.com/mockingio/engine/mock"
	"github.com/mockingio/engine/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("Load mock from YAML file", func(t *testing.T) {
		cfg, err := FromFile("fixtures/mock.yml")

		assert.True(t, cfg.Validate() == nil)

		var goldenFile = filepath.Join("fixtures", "mock.golden.yml")
		require.NoError(t, err)

		text, _ := yaml.Marshal(cfg)
		test.UpdateGoldenFile(t, goldenFile, text)

		assert.Equal(t, test.ReadGoldenFile(t, goldenFile), string(text))
	})

	t.Run("Load mock from YAML file, with ID generation option", func(t *testing.T) {
		mock, err := FromFile("fixtures/mock.yml", WithIDGeneration())
		require.NoError(t, err)

		assert.True(t, mock.ID != "")
		assert.True(t, mock.Routes[0].ID != "")
		assert.True(t, mock.Routes[0].Responses[0].ID != "")
		assert.True(t, mock.Routes[0].Responses[0].Rules[0].ID != "")
	})

	t.Run("When method, status is not presented, use default GET/200 as response", func(t *testing.T) {
		mock, err := FromFile("fixtures/mock_no_method_status.yml")
		require.NoError(t, err)

		assert.Equal(t, "GET", mock.Routes[0].Method)
		assert.Equal(t, 200, mock.Routes[0].Responses[0].Status)
	})

	t.Run("error loading config from YAML file", func(t *testing.T) {
		cfg, err := FromFile("")
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("error loading mock from empty yaml", func(t *testing.T) {
		cfg, err := FromYaml("")
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}
