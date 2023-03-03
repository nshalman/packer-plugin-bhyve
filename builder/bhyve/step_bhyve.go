package bhyve

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepBhyve struct {
	name string
}

func (step *stepBhyve) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	cd_device := fmt.Sprintf("2,ahci-cd,%s", state.Get("iso_path").(string))
	vnc_args := fmt.Sprintf("30:0,fbuf,vga=off,rfb=%s:%d,password=%s,wait",
		config.VNCBindAddress,
		state.Get("vnc_port").(int),
		state.Get("vnc_password").(string))

	args := []string{
		"-H",
		"-c", "1",
		"-l", "bootrom,/usr/share/bhyve/uefi-rom.bin",
		"-m", "1024",
		"-s", "0,hostbridge,model=i440fx",
		"-s", cd_device,
		"-s", vnc_args,
		"-s", "30:1,xhci,tablet",
		"-s", "31,lpc",
		step.name,
	}

	ui.Say(fmt.Sprintf("Starting bhyve VM %s", step.name))

	cmd := exec.Command("/usr/sbin/bhyve", args...)
	err := cmd.Start()
	if err != nil {
		err = fmt.Errorf("Error starting VM: %s", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (step *stepBhyve) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	vmarg := fmt.Sprintf("--vm=%s", step.name)

	args := []string{
		vmarg,
		"--destroy",
	}

	ui.Say(fmt.Sprintf("Stopping bhyve VM %s", step.name))

	cmd := exec.Command("/usr/sbin/bhyvectl", args...)
	err := cmd.Start()
	if err != nil {
		err = fmt.Errorf("Error stopping VM: %s", err)
	}
}
