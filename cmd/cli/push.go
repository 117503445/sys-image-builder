package main

import (
	"github.com/117503445/sys-image-builder/internal/pusher"
)

func push(cmd *CmdPush) error {
	return pusher.Push(pusher.Config{
		ImageRef: cmd.Image,
		Username: cmd.Username,
		Password: cmd.Password,
		Insecure: cmd.Insecure,
	})
}
