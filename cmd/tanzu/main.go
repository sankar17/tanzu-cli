// Copyright 2021-22 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"os/exec"

	"github.com/vmware-tanzu/tanzu-cli/pkg/command"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/log"
)

func main() {
	if err := command.Execute(); err != nil {
		if errStr, ok := err.(*exec.ExitError); ok {
			// If a plugin exited with an error, we don't want to print its
			// exit status as a string, but want to use it as our own exit code.
			os.Exit(errStr.ExitCode())
		} else {
			// We got an error other than a plugin exiting with an error, let's
			// print the error message.
			log.Fatal(err, "")
		}
	}
}
