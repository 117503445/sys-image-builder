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

// RegistryConfig holds the registry configuration.
type RegistryConfig struct {
	// ImageRef is the image reference (e.g., "registry:5000/repo:tag")
	ImageRef string

	// Username for registry authentication
	Username string

	// Password for registry authentication
	Password string

	// Insecure uses HTTP instead of HTTPS for the registry
	Insecure bool
}

// ImageConfig holds the container image configuration.
type ImageConfig struct {
	// Cmd is the default arguments to the entrypoint of the container.
	Cmd []string

	// Entrypoint is the entry point of the container.
	Entrypoint []string

	// WorkingDir is the current working directory of the container.
	WorkingDir string

	// User is the user that the container should run as.
	User string

	// ExposedPorts is a set of ports to expose from the container.
	ExposedPorts []string

	// Env is a list of environment variables in the form KEY=VALUE.
	Env []string

	// Labels are key-value pairs that are attached to the container.
	Labels map[string]string
}

// BuildConfig holds the build configuration.
type BuildConfig struct {
	// Excludes is a list of paths to exclude from the rootfs.
	// If empty, default excludes are used: /proc, /sys, /dev, /run, /tmp, /mnt, /media, /lost+found, /var/cache, /var/tmp
	Excludes []string
}

// Config holds the configuration for pushing a system image.
type Config struct {
	// RegistryConfig is the registry configuration.
	RegistryConfig RegistryConfig

	// BuildConfig is the build configuration.
	BuildConfig BuildConfig

	// ImageConfig is the container image configuration.
	ImageConfig ImageConfig
}

// Push packages the current system's rootfs into a container image and pushes it to the specified registry.
//
// The rootfs is packaged as a tar archive (excluding system directories by default) and
// pushed as a single layer container image.
func Push(ctx context.Context, cfg Config) error {
	return pusher.Push(ctx, pusher.Config{
		RegistryConfig: pusher.RegistryConfig{
			ImageRef: cfg.RegistryConfig.ImageRef,
			Username: cfg.RegistryConfig.Username,
			Password: cfg.RegistryConfig.Password,
			Insecure: cfg.RegistryConfig.Insecure,
		},
		BuildConfig: pusher.BuildConfig{
			Excludes: cfg.BuildConfig.Excludes,
		},
		ImageConfig: pusher.ImageConfig{
			Cmd:          cfg.ImageConfig.Cmd,
			Entrypoint:   cfg.ImageConfig.Entrypoint,
			WorkingDir:   cfg.ImageConfig.WorkingDir,
			User:         cfg.ImageConfig.User,
			ExposedPorts: cfg.ImageConfig.ExposedPorts,
			Env:          cfg.ImageConfig.Env,
			Labels:       cfg.ImageConfig.Labels,
		},
	})
}