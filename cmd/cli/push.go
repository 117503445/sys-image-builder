package main

import (
	"context"

	"github.com/117503445/sys-image-builder/pkg/sysimage"
	"github.com/rs/zerolog/log"
)

func push(cmd *CmdPush) error {
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)
	return sysimage.Push(ctx, sysimage.Config{
		RegistryConfig: sysimage.RegistryConfig{
			ImageRef: cmd.Image,
			Username: cmd.Username,
			Password: cmd.Password,
			Insecure: cmd.Insecure,
		},
	})
}
