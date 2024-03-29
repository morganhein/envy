package manager

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
	"path"
	"strings"

	"github.com/morganhein/envy/pkg/io"
)

type Manager interface {
	Start(ctx context.Context, config Recipe, operation Operation, name string) error
	// RunTask will explicitly only run a specified task, and will fail if it is not found
	//RunInstall will explicitly only run the installation of the package.
	//RunInstall(ctx context.Context, config Recipe, pkg string) error
}

type RunConfig struct {
	RecipeLocation string
	Operation      Operation
	Recipe         Recipe
	ForceInstaller string // ForceInstaller will try to force the specified installer
	Sudo           string // Sudo will force using sudo when performing commands
	Verbose        bool   // Talk more
	DryRun         bool   // Don't actually run installation/copy/symlink commands
	TargetDir      string // TargetDir is the base directory for symlinks, defaults to ${HOME}
	SourceDir      string // SourceDir is the base directory to search for source files to symlink against, defaults to dir(RecipeLocation)
	originalTask   string // used for environment variable replacement. Do we need?
}

type manager struct {
	d                 Decider
	r                 io.Shell
	dl                io.Downloader //not being used yet due to refactor
	fs                io.Filesystem
	updatedInstallers map[string]interface{}
}

func New(fs io.Filesystem, shell io.Shell) manager {
	d := NewDecider(shell)
	return manager{
		d:  d,
		r:  shell,
		fs: fs,
	}
}

// Start is the command line entrypoint
func (m *manager) Start(ctx context.Context, config RunConfig, name string) error {
	if m.updatedInstallers == nil {
		m.updatedInstallers = make(map[string]interface{})
	}
	tConfig, err := ResolveRecipe(m.fs, config.RecipeLocation)
	if err != nil {
		cobra.CheckErr(err)
	}
	config.Recipe = *tConfig
	io.PrintVerboseF(config.Verbose, "Operation: %v, Name: %v, verbose: %v, sudo: %v",
		config.Operation,
		name,
		config.Verbose,
		config.Sudo)
	if config.Operation == TASK {
		config.originalTask = name
		return m.RunTask(ctx, config, name)
	}
	if config.Operation == INSTALL {
		return m.RunInstall(ctx, config, name)
	}
	return xerrors.Errorf("Operation `%v` not supported", config.Operation)
}

func (m *manager) RunTask(ctx context.Context, config RunConfig, task string) error {
	//start tracking environment variables
	vars := envVariables{}
	hydrateEnvironment(config, vars)
	io.PrintVerbose(config.Verbose, fmt.Sprintf("original environment variables: %+v", vars), nil)
	return m.runTaskHelper(ctx, config, vars, task)
}

func (m *manager) RunInstall(ctx context.Context, config RunConfig, pkg string) error {
	//start tracking environment variables
	vars := envVariables{}
	hydrateEnvironment(config, vars)
	io.PrintVerbose(config.Verbose, fmt.Sprintf("original environment variables: %+v", vars), nil)
	//this should go straight to the pkg install helper, and none of this other business
	return m.installPkgHelper(ctx, config, vars, pkg)
}

func (m *manager) handleDependency(ctx context.Context, config RunConfig, vars envVariables, taskOrPkg string) error {
	if len(taskOrPkg) == 0 {
		return xerrors.New("task or package is empty")
	}
	// if the dependency is a task, run it
	if taskOrPkg[0] == '#' {
		return m.runTaskHelper(ctx, config, vars, taskOrPkg[1:])
	}
	//default is just a plain package name
	return m.installPkgHelper(ctx, config, vars, taskOrPkg)
}

/*runTaskHelper runs, in order:
* Determines if the installers required by the task are available
* If `run_if` passes
* If `skip_if` passes
* Downloads any necessary files
* Installs any deps
* Runs the pre_cmd commands
* Installs the package
* Runs the post_cmd commands
 */
func (m *manager) runTaskHelper(ctx context.Context, config RunConfig, vars envVariables, task string) error {
	io.PrintVerbose(config.Verbose, fmt.Sprintf("starting task [%v]", task), nil)
	//load the task
	t, ok := config.Recipe.Tasks[task]
	if !ok {
		return xerrors.Errorf("task '%v' not defined in config", task)
	}

	if sr := m.d.ShouldRun(ctx, t.SkipIf, t.RunIf); !sr {
		io.PrintVerbose(config.Verbose, fmt.Sprintf("task '%v' failed skip_if or run_if check", task), nil)
		return nil
	}

	//download the files
	for _, dlReq := range t.Download {
		if len(dlReq) != 2 {
			return xerrors.New("the download command must contain two parameters, the source and the target")
		}
		_, err := m.dl.Download(ctx, dlReq[0], dlReq[1])
		if err != nil {
			return err
		}
	}

	//run the deps
	for _, dep := range t.Deps {
		if err := m.handleDependency(ctx, config, vars, dep); err != nil {
			return err
		}
	}

	//run the pre-cmds
	for _, cmd := range t.PreCmds {
		if err := m.runCmdHelper(ctx, config, vars, cmd); err != nil {
			return err
		}
	}

	//install the packages
	for _, pkg := range t.Install {
		if err := m.RunInstall(ctx, config, pkg); err != nil {
			return err
		}
	}

	//run the post-cmds
	for _, cmd := range t.PostCmds {
		if err := m.runCmdHelper(ctx, config, vars, cmd); err != nil {
			return err
		}
	}

	return nil
}

// runCmdHelper runs any commands in pre/post cmds with variables replaced
func (m *manager) runCmdHelper(ctx context.Context, config RunConfig, vars envVariables, cmdLine string) error {
	//cleanup first
	cmdLine = strings.TrimSpace(cmdLine)
	sudo := determineSudo(config, nil)
	cmdLine = injectVars(cmdLine, vars, sudo)
	io.PrintVerbose(config.Verbose, fmt.Sprintf("running command `%v`", cmdLine), nil)
	out, err := m.r.Run(ctx, config.DryRun, cmdLine)
	io.PrintVerbose(config.Verbose, out, err)
	if err != nil {
		return err
	}
	return nil
}

func (m *manager) downloadHelper(ctx context.Context, dl Downloads) (string, error) {
	if len(dl) == 2 {
		return m.dl.Download(ctx, dl[0], dl[1])
	}
	return "", errors.New("incorrect syntax for a download command")
}

// TODO (@morgan): this should probably be removed? in lieu of the sync operation?
func (m *manager) symlinkHelper(ctx context.Context, config RunConfig, vars envVariables, link string) error {
	io.PrintVerbose(config.Verbose, fmt.Sprintf("creating symlink `%v`", link), nil)
	parts := strings.Split(link, " ")
	if len(parts) > 2 {
		return xerrors.New("unexpected symlink format, which is `from [to]`")
	}
	from := path.Join(config.SourceDir, parts[0])
	to := path.Join(config.TargetDir, parts[0])
	if len(parts) == 2 {
		to = path.Join(config.TargetDir, parts[1])
	}

	if config.DryRun {
		fmt.Printf("symlinking from %v to %v\n", from, to)
		return nil
	}

	out, err := func() (string, error) {
		fmt.Println(link)
		return "", nil
	}()
	io.PrintVerbose(config.Verbose, out, err)
	return err
}

func (m *manager) installPkgHelper(ctx context.Context, config RunConfig, vars envVariables, pkgName string) error {
	if len(pkgName) == 0 {
		return errors.New("unable to find the package name")
	}

	//look up the package in the config, if it exists.
	pkg := getPackage(config.Recipe, pkgName)
	io.PrintVerboseF(config.Verbose, "resolved package name to `%v`", pkg)
	//determine which installer is preferred with this package
	installer, err := determineBestAvailableInstaller(ctx, config, pkg, m.d)
	if err != nil {
		return err
	}
	io.PrintVerboseF(config.Verbose, "resolved installer to `%v`", installer.Name)

	//run the install commands for that installer
	//do we sudo, or do we not?
	sudo := determineSudo(config, installer)

	//insure installer has been updated, if possible
	if _, ok := m.updatedInstallers[installer.Name]; !ok && len(installer.Update) > 0 {
		cmdLine := replaceSudo(installer.Update, sudo)
		io.PrintVerboseF(config.Verbose, "running update for installer `%v` for the first time", installer.Name)
		_, err = m.r.Run(ctx, config.DryRun, cmdLine)
		if err != nil {
			return err
		}
		m.updatedInstallers[installer.Name] = nil
	}

	//determine package name in relation to the chosen installer
	newPkgName, ok := pkg[installer.Name]
	if !ok {
		newPkgName = pkgName
	}

	//TODO (@morgan): at this point, if it is a shell installer, call that instead

	cmdLine := installCommandVariableSubstitution(installer.Cmd, newPkgName, sudo)
	io.PrintVerboseF(config.Verbose, "running command `%v`", cmdLine)

	out, err := m.r.Run(ctx, config.DryRun, cmdLine)
	if err != nil {
		io.PrintVerbose(config.Verbose, out, err)
		return err
	}
	io.PrintVerboseF(config.Verbose, "package installation successful")
	return nil
}
