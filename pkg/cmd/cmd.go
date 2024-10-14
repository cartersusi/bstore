package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func RunCMD_fs(fs string) error {
	cmd := exec.Command("bash", "-c", fs)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
	}

	return nil
}

func RunCMD_fs_loose(fs string) {
	cmd := exec.Command("bash", "-c", fs)
	_ = cmd.Run()
	return
}

func RunCMD(name string, arg ...string) error {
	fs := name + " " + strings.Join(arg, " ")
	fmt.Printf("Running %s\n", fs)

	cmd := exec.Command("bash", "-c", fs)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
	}

	return nil
}

func GetCMD(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
