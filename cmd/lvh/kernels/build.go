// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package kernels

import (
	"context"

	"github.com/cilium/little-vm-helper/pkg/kernels"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func buildCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build <kernel>",
		Short: "build kernel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logrus.New()
			kname := args[0]
			return kernels.BuildKernel(context.Background(), log, dirName, kname, false /* TODO: add fetch flag */)
		},
	}
}
