package builder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

type BuildConfig struct {
	Command string `hcl:"command"`
	SourceDir string `hcl:"source_dir"`
	OutputDir string `hcl:"output_dir"`
	FileName string `hcl:"file_name"`
}

type Builder struct {
	config BuildConfig
}

// Implement Configurable
func (b *Builder) Config() (interface{}, error) {
	return &b.config, nil
}

// Implement ConfigurableNotify
func (b *Builder) ConfigSet(config interface{}) error {
	c, ok := config.(*BuildConfig)
	if !ok {
		// The Waypoint SDK should ensure this never gets hit
		return fmt.Errorf("Expected *BuildConfig as parameter")
	}

	if c.Command != "assembly" && c.Command != "build" {
		return fmt.Errorf("sbt assembly and sbt build are the only support commands")
	}
	// validate the config
	//if c.Directory == "" {
	//	return fmt.Errorf("Directory must be set to a valid directory")
	//}

	return nil
}

// Implement Builder
func (b *Builder) BuildFunc() interface{} {
	// return a function which will be called by Waypoint
	return b.build
}

// A BuildFunc does not have a strict signature, you can define the parameters
// you need based on the Available parameters that the Waypoint SDK provides.
// Waypoint will automatically inject parameters as specified
// in the signature at run time.
//
// Available input parameters:
// - context.Context
// - *component.Source
// - *component.JobInfo
// - *component.DeploymentConfig
// - *datadir.Project
// - *datadir.App
// - *datadir.Component
// - hclog.Logger
// - terminal.UI
// - *component.LabelSet
//
// The output parameters for BuildFunc must be a Struct which can
// be serialzied to Protocol Buffers binary format and an error.
// This Output Value will be made available for other functions
// as an input parameter.
// If an error is returned, Waypoint stops the execution flow and
// returns an error to the user.
func (b *Builder) build(ctx context.Context, ui terminal.UI) (*Binary, error) {
	u := ui.Status()
	defer u.Close()
	u.Update("Building application")



	outputPath, err := filepath.Abs(b.config.OutputDir + "/" + b.config.FileName)

	if err != nil {
		u.Step(terminal.ErrorStyle, "Error finding output path: " + err.Error())
	}

	outputArg := fmt.Sprintf("set assemblyOutputPath in assembly := new File(\"%s\")", outputPath)

	c := exec.Command(
		"sbt",
		outputArg,
		b.config.Command,
	)
	c.Dir = b.config.SourceDir

	_, w := io.Pipe()
	defer w.Close()
	c.Stdout = w

	var b2 bytes.Buffer
	c.Stdout = &b2

	io.Copy(os.Stdout, &b2)


	err = c.Run()
	c.Wait()



	for _, line := range strings.Split(b2.String(), "\n") {
		if err != nil {
			u.Step(terminal.ErrorStyle, line)
		} else {
			u.Step(terminal.StatusOK, line)
		}
	}

	if err != nil{
		return nil, err
	}

	return &Binary{ Path: b.config.OutputDir + "/" + b.config.FileName }, nil
}
