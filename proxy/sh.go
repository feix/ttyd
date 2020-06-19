package proxy

import (
	"context"
	"io"
	"os/exec"

	"github.com/feix/ttyd/config"
)

func SH(ctx context.Context, rw io.ReadWriter) error {
	cmd := exec.CommandContext(ctx, config.GetEnv("SHELL", "/bin/bash"))
	cmd.Stdout = rw
	cmd.Stderr = rw
	cmd.Stdin = rw
	return cmd.Run()
}
