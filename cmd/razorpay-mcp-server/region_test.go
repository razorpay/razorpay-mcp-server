package main

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/razorpay/razorpay-mcp-server/pkg/region"
)

func TestRegionFlags(t *testing.T) {
	// Other tests call viper.Reset, which drops bindings made in init.
	bindRegionConfig()

	t.Run("region flags are registered", func(t *testing.T) {
		assert.NotNil(t, rootCmd.PersistentFlags().Lookup("region"))
		assert.NotNil(t, rootCmd.PersistentFlags().Lookup("region-config"))
	})

	t.Run("RAZORPAY_REGION env sets the region", func(t *testing.T) {
		t.Setenv("RAZORPAY_REGION", "SG")
		assert.Equal(t, "SG", viper.GetString("razorpay_region"))
	})

	t.Run("a bare REGION env is ignored", func(t *testing.T) {
		t.Setenv("RAZORPAY_REGION", "")
		t.Setenv("REGION", "us-east-1")
		assert.Empty(t, viper.GetString("razorpay_region"))
	})

	t.Run("RAZORPAY_REGION_CONFIG env sets the config path",
		func(t *testing.T) {
			t.Setenv("RAZORPAY_REGION_CONFIG", "/tmp/regions.yaml")
			assert.Equal(t, "/tmp/regions.yaml",
				viper.GetString("razorpay_region_config"))
		})
}

func TestBuildStdioDeps(t *testing.T) {
	t.Run("no region uses IN and the India API host", func(t *testing.T) {
		client, profile, err := buildStdioDeps("key", "secret", "", "")
		require.NoError(t, err)
		assert.Equal(t, "IN", profile.CountryCode)
		assert.Equal(t, "https://api.razorpay.com", client.Request.BaseURL)
	})

	t.Run("SG uses the SG API host", func(t *testing.T) {
		client, profile, err := buildStdioDeps("key", "secret", "sg", "")
		require.NoError(t, err)
		assert.Equal(t, "SG", profile.CountryCode)
		assert.Equal(t, "https://api-sg.razorpay.com", client.Request.BaseURL)
	})

	t.Run("unknown region fails", func(t *testing.T) {
		_, _, err := buildStdioDeps("key", "secret", "XX", "")
		assert.ErrorContains(t, err, `no region profile for "XX"`)
	})

	t.Run("unreadable region config fails", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing.yaml")
		_, _, err := buildStdioDeps("key", "secret", "", missing)
		assert.ErrorContains(t, err, "region config")
	})
}

func TestRunStdioServerUsesProfileFromContext(t *testing.T) {
	ctx, cancel, obs, client := setupTestServer(t)
	defer cancel()

	profiles, err := region.Load("")
	require.NoError(t, err)
	ctx = region.WithProfile(ctx, profiles["SG"])

	err = runStdioServer(ctx, obs, client, []string{"payouts"}, false)
	assert.ErrorContains(t, err,
		`toolset "payouts" has no tools available in region SG`)
}
