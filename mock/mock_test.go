package mock_test

import (
	_ "embed"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"testing"

	. "github.com/mockingio/engine/mock"
	"github.com/mockingio/engine/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("Load config from YAML file", func(t *testing.T) {
		cfg, err := FromFile("fixtures/mock.yml")

		assert.True(t, cfg.Validate() == nil)

		var goldenFile = filepath.Join("fixtures", "mock.golden.yml")
		require.NoError(t, err)

		text, _ := yaml.Marshal(cfg)
		test.UpdateGoldenFile(t, goldenFile, text)

		assert.Equal(t, test.ReadGoldenFile(t, goldenFile), string(text))
	})

	t.Run("Load config from JSON file", func(t *testing.T) {
		cfg, err := FromFile("fixtures/mock.json")
		require.NoError(t, err)

		require.NoError(t, cfg.Validate())

		var goldenFile = filepath.Join("fixtures", "mock.golden.json")
		require.NoError(t, err)

		text, _ := json.MarshalIndent(cfg, "", "  ")

		test.UpdateGoldenFile(t, goldenFile, text)

		assert.Equal(t, test.ReadGoldenFile(t, goldenFile), string(text))
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
