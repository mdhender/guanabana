// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var cmdRoot = &cobra.Command{
		Use:   "guanabana",
		Short: "Guanabana Parser Generator",
		Long:  `Yet another LALR parser generator.`,
	}
	if cmd, err := cmdBuild(); err != nil {
		log.Fatalf("build: %v\n", err)
	} else {
		cmdRoot.AddCommand(cmd)
	}
	if cmd, err := cmdDump(); err != nil {
		log.Fatalf("dump: %v\n", err)
	} else {
		cmdRoot.AddCommand(cmd)
	}
	if cmd, err := cmdParse(); err != nil {
		log.Fatalf("parse: %v\n", err)
	} else {
		cmdRoot.AddCommand(cmd)
	}

	if err := cmdRoot.Execute(); err != nil {
		os.Exit(1)
	}
}

func cmdBuild() (*cobra.Command, error) {
	var cmd = &cobra.Command{
		Use:   "build",
		Short: "build",
	}
	return cmd, nil
}

func cmdDump() (*cobra.Command, error) {
	var cmd = &cobra.Command{
		Use:   "dump",
		Short: "dump",
	}
	return cmd, nil
}

func cmdParse() (*cobra.Command, error) {
	var cmd = &cobra.Command{
		Use:   "parse",
		Short: "parse",
	}
	return cmd, nil
}
