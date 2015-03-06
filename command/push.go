package command

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/atlas-go/archive"
)

type PushCommand struct {
	Meta

	// client is the client to use for the actual push operations.
	// If this isn't set, then the Atlas client is used. This should
	// really only be set for testing reasons (and is hence not exported).
	client pushClient
}

func (c *PushCommand) Run(args []string) int {
	var atlasToken string
	var moduleLock bool
	args = c.Meta.process(args, false)
	cmdFlags := flag.NewFlagSet("push", flag.ContinueOnError)
	cmdFlags.StringVar(&c.Meta.statePath, "state", DefaultStateFilename, "path")
	cmdFlags.StringVar(&atlasToken, "token", "", "")
	cmdFlags.BoolVar(&moduleLock, "module-lock", true, "")
	cmdFlags.Usage = func() { c.Ui.Error(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	// The pwd is used for the configuration path if one is not given
	pwd, err := os.Getwd()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error getting pwd: %s", err))
		return 1
	}

	// Get the path to the configuration depending on the args.
	var configPath string
	args = cmdFlags.Args()
	if len(args) > 1 {
		c.Ui.Error("The apply command expects at most one argument.")
		cmdFlags.Usage()
		return 1
	} else if len(args) == 1 {
		configPath = args[0]
	} else {
		configPath = pwd
	}

	// Verify the state is remote, we can't push without a remote state
	s, err := c.State()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to read state: %s", err))
		return 1
	}
	if !s.State().IsRemote() {
		c.Ui.Error(
			"Remote state is not enabled. For Atlas to run Terraform\n" +
				"for you, remote state must be used and configured. Remote\n" +
				"state via any backend is accepted, not just Atlas. To\n" +
				"configure remote state, use the `terraform remote config`\n" +
				"command.")
		return 1
	}

	// Build the context based on the arguments given
	ctx, planned, err := c.Context(contextOpts{
		Path:      configPath,
		StatePath: c.Meta.statePath,
	})
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	if planned {
		c.Ui.Error(
			"A plan file cannot be given as the path to the configuration.\n" +
				"A path to a module (directory with configuration) must be given.")
		return 1
	}

	// Ask for input
	if err := ctx.Input(c.InputMode()); err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Error while asking for variable input:\n\n%s", err))
		return 1
	}

	// Build the archiving options, which includes everything it can
	// by default according to VCS rules but forcing the data directory.
	archiveOpts := &archive.ArchiveOpts{
		VCS: true,
		Extra: map[string]string{
			DefaultDataDir: c.DataDir(),
		},
	}
	if !moduleLock {
		// If we're not locking modules, then exclude the modules dir.
		archiveOpts.Exclude = append(
			archiveOpts.Exclude,
			filepath.Join(c.DataDir(), "modules"))
	}

	archiveR, err := archive.CreateArchive(configPath, archiveOpts)
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"An error has occurred while archiving the module for uploading:\n"+
				"%s", err))
		return 1
	}

	// Upsert!
	if err := c.client.Upsert(archiveR, archiveR.Size); err != nil {
		c.Ui.Error(fmt.Sprintf(
			"An error occurred while uploading the module:\n\n%s", err))
		return 1
	}

	return 0
}

func (c *PushCommand) Help() string {
	helpText := `
Usage: terraform push [options] [DIR]

  Upload this Terraform module to an Atlas server for remote
  infrastructure management.

Options:

  -module-lock=true    If true (default), then the modules are locked at
                       their current checkout and uploaded completely. This
                       prevents Atlas from running "terraform get".

  -token=<token>       Access token to use to upload. If blank, the ATLAS_TOKEN
                       environmental variable will be used.

`
	return strings.TrimSpace(helpText)
}

func (c *PushCommand) Synopsis() string {
	return "Upload this Terraform module to Atlas to run"
}

// pushClient is implementd internally to control where pushes go. This is
// either to Atlas or a mock for testing.
type pushClient interface {
	Upsert(io.Reader, int64) error
}

type mockPushClient struct {
	File string

	UpsertCalled bool
	UpsertError  error
}

func (c *mockPushClient) Upsert(data io.Reader, size int64) error {
	f, err := os.Create(c.File)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.CopyN(f, data, size); err != nil {
		return err
	}

	c.UpsertCalled = true
	return c.UpsertError
}