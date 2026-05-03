package main

import (
	"context"

	"github.com/117503445/sys-image-builder/pkg/sysimage"
	"github.com/rs/zerolog/log"
)

// push 执行系统镜像推送。
//
// 参数 cmd 是 push 命令参数。
func push(cmd *CmdPush) error {
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)

	excludes := append([]string{}, cmd.Exclude...)
	if cmd.DefaultExcludes {
		excludes = append(sysimage.DefaultExcludes(), excludes...)
	}

	return sysimage.Push(ctx, sysimage.Config{
		RegistryConfig: sysimage.RegistryConfig{
			ImageRef: cmd.Image,
			Username: cmd.Username,
			Password: cmd.Password,
			Insecure: cmd.Insecure,
		},
		BuildConfig: sysimage.BuildConfig{
			Excludes: excludes,
		},
	})
}
