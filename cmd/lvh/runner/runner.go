// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package runner

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cilium/little-vm-helper/pkg/runner"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

var (
	rcnf RunConf

	ports []string
)

func RunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "run",
		Short:        "run/start VMs based on generated base images and kernels",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			rcnf.Logger = logrus.New()

			rcnf.ForwardedPorts, err = runner.ParsePortForward(ports)
			if err != nil {
				return fmt.Errorf("Port flags: %w", err)
			}

			t0 := time.Now()

			err = StartQemu(rcnf)
			dur := time.Since(t0).Round(time.Millisecond)
			fmt.Printf("Execution took %v\n", dur)
			if err != nil {
				return fmt.Errorf("Qemu exited with an error: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&rcnf.Image, "image", "", "VM image file path")
	cmd.MarkFlagRequired("image")
	cmd.Flags().StringVar(&rcnf.KernelFname, "kernel", "", "kernel filename to boot with. (if empty no -kernel option will be passed to qemu)")
	cmd.Flags().BoolVar(&rcnf.QemuPrint, "qemu-cmd-print", false, "Do not run the qemu command, just print it")
	cmd.Flags().BoolVar(&rcnf.DisableKVM, "qemu-disable-kvm", false, "Do not use KVM acceleration, even if /dev/kvm exists")
	cmd.Flags().BoolVar(&rcnf.Daemonize, "daemonize", false, "daemonize QEMU after initializing")
	cmd.Flags().StringVar(&rcnf.HostMount, "host-mount", "", "Mount the specified host directory in the VM using a 'host_mount' tag")
	cmd.Flags().StringArrayVarP(&ports, "port", "p", nil, "Forward a port (hostport[:vmport[:tcp|udp]])")
	cmd.Flags().IntVar(&rcnf.SerialPort, "serial-port", 0, "Port for serial console")
	cmd.Flags().IntVar(&rcnf.CPU, "cpu", 2, "CPU count (-smp)")
	cmd.Flags().StringVar(&rcnf.Mem, "mem", "4G", "RAM size (-m)")
	cmd.Flags().StringVar(&rcnf.CPUKind, "cpu-kind", "kvm64", "CPU kind to use (-cpu), has no effect when KVM is disabled")
	cmd.Flags().IntVar(&rcnf.QemuMonitorPort, "qemu-monitor-port", 0, "Port for QEMU monitor")
	cmd.Flags().BoolVarP(&rcnf.Verbose, "verbose", "v", false, "Print qemu command before running it")

	return cmd
}

const qemuBin = "qemu-system-x86_64"

func StartQemu(rcnf RunConf) error {
	qemuArgs, err := BuildQemuArgs(rcnf.Logger, &rcnf)
	if err != nil {
		return err
	}

	if rcnf.QemuPrint || rcnf.Verbose {
		var sb strings.Builder
		sb.WriteString(qemuBin)
		for _, arg := range qemuArgs {
			sb.WriteString(" ")
			if len(arg) > 0 && arg[0] == '-' {
				sb.WriteString("\\\n\t")
			}
			sb.WriteString(arg)
		}

		fmt.Printf("%s\n", sb.String())
		// We don't want to return early if running in verbose mode
		if rcnf.QemuPrint {
			return nil
		}
	}

	qemuPath, err := exec.LookPath(qemuBin)
	if err != nil {
		return err
	}

	return unix.Exec(qemuPath, append([]string{qemuBin}, qemuArgs...), nil)
}
