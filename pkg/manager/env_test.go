package manager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAndReadConfig(t *testing.T) {
	config, err := LoadPackageConfig(context.Background(), "../packages.toml")
	assert.NoError(t, err)
	assert.NotNil(t, config)
}

func TestLoadDefaultInstallers(t *testing.T) {
	installers, err := loadDefaultInstallers(Config{})
	assert.NoError(t, err)
	assert.NotNil(t, installers)
}

func TestCombineInstallers(t *testing.T) {
	c := Config{
		Installers: map[string]Installer{
			"TEST": {
				Name:   "TEST_PKG_MANAGER",
				SkipIf: nil,
				RunIf:  []string{"which ls"}, //assumed that LS exists pretty much everywhere
				Sudo:   false,
			},
		},
	}
	installers, err := loadDefaultInstallers(c)
	assert.NoError(t, err)
	assert.NotNil(t, installers)
	assert.Equal(t, "TEST_PKG_MANAGER", c.Installers["TEST"].Name)
}