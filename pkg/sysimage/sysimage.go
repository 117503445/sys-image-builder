// Package sysimage provides functionality to build and push system images to container registries.
//
// This package allows Go programs to integrate system image building capabilities,
// enabling them to package the current system's rootfs into a container image and
// push it to a container registry.
//
// Example usage:
//
//	err := sysimage.Push(sysimage.Config{
//	    ImageRef: "registry.example.com/my-image:v1",
//	    Username: "user",
//	    Password: "pass",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
package sysimage

import (
	"context"

	"github.com/117503445/sys-image-builder/internal/pusher"
)

// Config holds the configuration for pushing a system image.
type Config struct {
	// ImageRef is the image reference (e.g., "registry:5000/repo:tag")
	ImageRef string

	// Username for registry authentication
	Username string

	// Password for registry authentication
	Password string

	// Insecure uses HTTP instead of HTTPS for the registry
	Insecure bool

	// Excludes is a list of paths to exclude from the rootfs.
	// If empty, default excludes are used: /proc, /sys, /dev, /run, /tmp, /mnt, /media, /lost+found, /var/cache, /var/tmp
	Excludes []string
}

// Push packages the current system's rootfs into a container image and pushes it to the specified registry.
//
// The rootfs is packaged as a tar archive (excluding system directories by default) and
// pushed as a single layer container image.
func Push(ctx context.Context, cfg Config) error {
	return pusher.Push(ctx, pusher.Config{
		ImageRef: cfg.ImageRef,
		Username: cfg.Username,
		Password: cfg.Password,
		Insecure: cfg.Insecure,
		Excludes: cfg.Excludes,
	})
}