// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/benchmarks/sweet/common/log"
)

type Go struct {
	Tool       string
	Env        *Env
	PassOutput bool
}

func SystemGoTool() (*Go, error) {
	tool, err := exec.LookPath("go")
	if err != nil {
		return nil, fmt.Errorf("looking for system Go: %w", err)
	}
	return &Go{
		Tool: tool,
		Env:  NewEnvFromEnviron().MustSet("GOROOT=" + filepath.Dir(filepath.Dir(tool))),
	}, nil
}

func (g *Go) Do(args ...string) error {
	cmd := exec.Command(g.Tool, args...)
	cmd.Env = g.Env.Collapse()
	if g.PassOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	log.TraceCommand(cmd, false)
	if g.PassOutput {
		return cmd.Run()
	}
	// Use cmd.Output to get an ExitError with Stderr populated.
	_, err := cmd.Output()
	return err
}

func (g *Go) List(args ...string) ([]byte, error) {
	cmd := exec.Command(g.Tool, append([]string{"list"}, args...)...)
	cmd.Env = g.Env.Collapse()
	log.TraceCommand(cmd, false)
	return cmd.Output()
}

func (g *Go) GOROOT() string {
	return filepath.Dir(filepath.Dir(g.Tool))
}

func (g *Go) BuildPackage(pkg, out string) error {
	if pkg[0] == '/' || pkg[0] == '.' {
		return fmt.Errorf("path used as package in go build")
	}
	return g.Do("build", "-o", out, pkg)
}

func (g *Go) BuildPath(path, out string) error {
	if path[0] != '/' && path[0] != '.' {
		path = "./" + path
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	defer chdir(cwd)
	if err := chdir(path); err != nil {
		return fmt.Errorf("failed to enter build directory: %w", err)
	}
	return g.Do("build", "-o", out)
}

func chdir(path string) error {
	log.CommandPrintf("cd %s", path)
	return os.Chdir(path)
}
