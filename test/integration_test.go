//go:build integrated
// +build integrated

package test

import (
	"fmt"
	"github.com/morganhein/envy/pkg/io"
	"github.com/morganhein/envy/pkg/manager"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

// Puts the config file in /usr/var/envy/default.toml and defaults are loaded
func TestLoadConfigFromUsrDefault(t *testing.T) {
	defaultLocation := "/usr/share/envy/default.toml"
	err := os.Mkdir("/usr/share/envy", os.ModeDir)
	assert.NoError(t, err)
	_, err = copy("/app/configs/default.toml", defaultLocation)
	assert.NoError(t, err)
	e, err := exists(defaultLocation)
	assert.NoError(t, err)
	assert.True(t, e)

	r, err := manager.ResolveRecipe(io.NewFilesystem(), "")
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Contains(t, r.InstallerDefs, "apt")
}

// Puts config in $HOME/.config/envy/default.toml and defaults are loaded
func TestLoadDefaultConfigFromHomeConfig(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	assert.NoError(t, err)
	homeConfigLocation := fmt.Sprintf("%v/.config/envy/default.toml", homeDir)
	err = os.MkdirAll(fmt.Sprintf("%v/.config/envy/", homeDir), os.ModeDir)
	assert.NoError(t, err)
	_, err = copy("../configs/default.toml", homeConfigLocation)
	assert.NoError(t, err)
	e, err := exists(homeConfigLocation)
	assert.NoError(t, err)
	assert.True(t, e)

	r, err := manager.ResolveRecipe(io.NewFilesystem(), "")
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Contains(t, r.InstallerDefs, "apt")
}

func TestWhich(t *testing.T) {
	r, err := io.CreateShell()
	assert.NoError(t, err)
	ctx, cancel := newCtx(10 * time.Second)
	//assert we get a known positive
	exists, out, err := r.Which(ctx, "ls")
	cancel()
	assert.NoError(t, err, out)
	assert.True(t, true)

	//assert we get a known negative
	exists, out, err = r.Which(ctx, "monkey-pox-and-covid-suck")
	cancel()
	assert.Error(t, err, out)
	assert.False(t, exists)
}

func TestInstallCommandInstallsPackage(t *testing.T) {
	sh, err := io.CreateShell()
	assert.NoError(t, err)
	ctx, cancel := newCtx(10 * time.Second)
	//assert vim doesn't already exist
	exists, out, err := sh.Which(ctx, "vim")
	cancel()
	assert.Error(t, err, out)
	assert.False(t, exists)

	//install it
	ctx, cancel = newCtx(10 * time.Second)
	mgr := manager.New(io.NewFilesystem(), sh)
	appConfig := manager.RunConfig{
		RecipeLocation: "/app/configs/default.toml",
		Operation:      manager.INSTALL,
		Sudo:           "false",
		Verbose:        false,
	}
	err = mgr.Start(ctx, appConfig, "vim")
	cancel()
	assert.NoError(t, err)

	//assert vim exists
	exists, out, err = sh.Which(ctx, "vim")
	cancel()
	assert.NoError(t, err)
	assert.True(t, exists, out)
}

func TestTaskInstallsPackageCorrectly(t *testing.T) {
	//copy default installers first
	defaultLocation := "/usr/share/envy/default.toml"
	_, err := copy("/app/configs/default.toml", defaultLocation)
	assert.NoError(t, err)

	// make shell
	sh, err := io.CreateShell()
	assert.NoError(t, err)
	ctx, cancel := newCtx(10 * time.Second)

	//assert vim doesn't already exist
	exists, out, err := sh.Which(ctx, "vim")
	cancel()
	assert.Error(t, err, out)
	assert.False(t, exists)

	//install it
	ctx, cancel = newCtx(10 * time.Second)
	mgr := manager.New(io.NewFilesystem(), sh)
	appConfig := manager.RunConfig{
		RecipeLocation: "/app/test/configs/task.toml",
		Operation:      manager.TASK,
		Sudo:           "false",
		Verbose:        false,
	}
	err = mgr.Start(ctx, appConfig, "vim")
	cancel()
	assert.NoError(t, err)

	//assert vim exists
	exists, out, err = sh.Which(ctx, "vim")
	cancel()
	assert.NoError(t, err)
	assert.True(t, exists, out)
}

// This test assumes the "vim" task has a dependency on "nano",
// so therefore should install nano as well
func TestTaskInstallsPkgDepsCorrectly(t *testing.T) {
	//copy default installers first
	defaultLocation := "/usr/share/envy/default.toml"
	_, err := copy("/app/configs/default.toml", defaultLocation)
	assert.NoError(t, err)

	// make shell
	sh, err := io.CreateShell()
	assert.NoError(t, err)
	ctx, cancel := newCtx(10 * time.Second)

	//assert nano doesn't already exist
	exists, out, err := sh.Which(ctx, "nano")
	cancel()
	assert.Error(t, err, out)
	assert.False(t, exists)

	//install it
	ctx, cancel = newCtx(10 * time.Second)
	mgr := manager.New(io.NewFilesystem(), sh)
	appConfig := manager.RunConfig{
		RecipeLocation: "/app/test/configs/task_with_pkg_deps.toml",
		Operation:      manager.TASK,
		Sudo:           "false",
		Verbose:        false,
	}
	err = mgr.Start(ctx, appConfig, "vim")
	cancel()
	assert.NoError(t, err)

	//assert nano exists
	exists, out, err = sh.Which(ctx, "nano")
	cancel()
	assert.NoError(t, err)
	assert.True(t, exists, out)
}

func TestTaskInstallsTaskDepsCorrectly(t *testing.T) {
	//copy default installers first
	defaultLocation := "/usr/share/envy/default.toml"
	_, err := copy("/app/configs/default.toml", defaultLocation)
	assert.NoError(t, err)

	// shell
	sh, err := io.CreateShell()
	assert.NoError(t, err)
	ctx, cancel := newCtx(10 * time.Second)

	//assert parted doesn't already exist
	exists, out, err := sh.Which(ctx, "parted")
	cancel()
	assert.Error(t, err, out)
	assert.False(t, exists)

	//install it
	ctx, cancel = newCtx(10 * time.Second)
	mgr := manager.New(io.NewFilesystem(), sh)
	appConfig := manager.RunConfig{
		RecipeLocation: "/app/test/configs/task_with_task_deps.toml",
		Operation:      manager.TASK,
		Sudo:           "false",
		Verbose:        false,
	}
	err = mgr.Start(ctx, appConfig, "vim")
	cancel()
	assert.NoError(t, err)

	//assert parted exists
	exists, out, err = sh.Which(ctx, "parted")
	cancel()
	assert.NoError(t, err)
	assert.True(t, exists, out)
}

func TestTaskPreCmd(t *testing.T) {
	//copy default installers first
	defaultLocation := "/usr/share/envy/default.toml"
	_, err := copy("/app/configs/default.toml", defaultLocation)
	assert.NoError(t, err)

	// shell
	sh, err := io.CreateShell()
	assert.NoError(t, err)
	ctx, cancel := newCtx(10 * time.Second)

	// filesystem
	fs := io.NewFilesystem()

	//assert the /tmp/pre_cmd file does not exist
	_, err = fs.Stat("/tmp/pre_cmd")
	assert.Error(t, err)

	//install it
	ctx, cancel = newCtx(10 * time.Second)
	mgr := manager.New(fs, sh)
	appConfig := manager.RunConfig{
		RecipeLocation: "/app/test/configs/task_with_pre_cmd.toml",
		Operation:      manager.TASK,
		Sudo:           "false",
		Verbose:        false,
	}
	err = mgr.Start(ctx, appConfig, "vim")
	cancel()
	assert.NoError(t, err)

	//assert the /tmp/pre_cmd file exists
	_, err = fs.Stat("/tmp/pre_cmd")
	assert.NoError(t, err)
}

func TestTaskPostCmd(t *testing.T) {
	//copy default installers first
	defaultLocation := "/usr/share/envy/default.toml"
	_, err := copy("/app/configs/default.toml", defaultLocation)
	assert.NoError(t, err)

	// shell
	sh, err := io.CreateShell()
	assert.NoError(t, err)
	ctx, cancel := newCtx(10 * time.Second)

	// filesystem
	fs := io.NewFilesystem()

	//assert the /tmp/post_cmd file does not exist
	_, err = fs.Stat("/tmp/post_cmd")
	assert.Error(t, err)

	//install it
	ctx, cancel = newCtx(10 * time.Second)
	mgr := manager.New(fs, sh)
	appConfig := manager.RunConfig{
		RecipeLocation: "/app/test/configs/task_with_post_cmd.toml",
		Operation:      manager.TASK,
		Sudo:           "false",
		Verbose:        false,
	}
	err = mgr.Start(ctx, appConfig, "vim")
	cancel()
	assert.NoError(t, err)

	//assert the /tmp/post_cmd file exists
	_, err = fs.Stat("/tmp/post_cmd")
	assert.NoError(t, err)
}

func TestTaskSpecificInstaller(t *testing.T) {
	//copy default installers first
	defaultLocation := "/usr/share/envy/default.toml"
	_, err := copy("/app/configs/default.toml", defaultLocation)
	assert.NoError(t, err)

	// shell
	sh, err := io.CreateShell()
	assert.NoError(t, err)
	ctx, cancel := newCtx(10 * time.Second)

	// filesystem
	fs := io.NewFilesystem()

	//assert the /tmp/post_cmd file does not exist
	_, err = fs.Stat("/tmp/post_cmd")
	assert.Error(t, err)

	//install it
	ctx, cancel = newCtx(10 * time.Second)
	mgr := manager.New(fs, sh)
	appConfig := manager.RunConfig{
		RecipeLocation: "/app/test/configs/task_with_post_cmd.toml",
		Operation:      manager.TASK,
		Sudo:           "false",
		Verbose:        false,
	}
	err = mgr.Start(ctx, appConfig, "vim")
	cancel()
	assert.NoError(t, err)

	//assert the /tmp/post_cmd file exists
	_, err = fs.Stat("/tmp/post_cmd")
	assert.NoError(t, err)
}
