package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/craiggwilson/goke/task"
	"github.com/craiggwilson/goke/task/command"
)

var packages = []string{
	"./analyzer",
	"./ast",
	"./astprint",
	"./eval",
	"./internal/bsonutil",
	"./internal/decimalutil",
	"./normalizer",
	"./optimizer",
	"./parser",
}
var registry = task.NewRegistry()

func init() {
	// Fmt
	registry.Declare("fmt").Description("formats the files").Do(func(ctx *task.Context) error {
		args := []string{"-s", "-l", "-w"}
		if ctx.Verbose {
			args = append(args, "-e")
		}
		args = append(args, packages...)

		return command.Command("gofmt", args...)(ctx)
	})

	// Static-Analysis
	registry.Declare("sa").Description("runs static analysis").DependsOn("sa-fmt", "sa-lint")
	registry.Declare("sa-fmt").Description("runs the formatter").Do(func(ctx *task.Context) error {
		args := []string{"-s", "-l"}
		if ctx.Verbose {
			args = append(args, "-e")
		}
		args = append(args, packages...)

		cmd := exec.CommandContext(ctx, "gofmt", args...)

		if !ctx.DryRun {
			output, err := cmd.CombinedOutput()
			if err != nil {
				return err
			}

			if len(output) > 0 {
				_, _ = ctx.Write(output)
				return fmt.Errorf("some files are not formatted according to gofmt")
			}
		}

		return nil
	})

	registry.Declare("sa-lint").Description("runs the linter").Do(func(ctx *task.Context) error {
		args := []string{"run", "--deadline", "10m"}
		args = append(args, packages...)
		return command.Command("golangci-lint", args...)(ctx)
	})

	// Test
	registry.Declare("test").Description("runs all tests").DependsOn("test-unit")
	registry.Declare("test-unit").Description("runs unit tests").Do(func(ctx *task.Context) error {
		args := []string{"test", "-count=1"}
		if ctx.Verbose {
			args = append(args, "-v")
		}
		args = append(args, packages...)
		return command.Command("go", args...)(ctx)
	})

	// Verify
	registry.Declare("verify").Description("verifies all the code").DependsOn("sa", "test")
}

func main() {
	err := task.Run(registry, os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(1)
		}
		fmt.Println(err)
		os.Exit(2)
	}
}
