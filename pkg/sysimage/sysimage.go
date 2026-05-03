// Package sysimage 提供系统 rootfs 打包和推送能力。
package sysimage

import (
	"context"
	"io"

	"github.com/117503445/sys-image-builder/internal/pusher"
)

// RegistryConfig 表示镜像仓库配置。
type RegistryConfig struct {
	// ImageRef 是镜像引用，例如 "registry:5000/repo:tag"。
	ImageRef string

	// Username 是镜像仓库认证用户名。
	Username string

	// Password 是镜像仓库认证密码。
	Password string

	// Insecure 表示使用 HTTP 访问镜像仓库。
	Insecure bool
}

// ImageConfig 表示容器镜像配置。
type ImageConfig struct {
	// Cmd 是容器入口的默认参数。
	Cmd []string

	// Entrypoint 是容器入口命令。
	Entrypoint []string

	// WorkingDir 是容器工作目录。
	WorkingDir string

	// User 是容器运行用户。
	User string

	// ExposedPorts 是容器暴露端口列表。
	ExposedPorts []string

	// Env 是 KEY=VALUE 格式的环境变量列表。
	Env []string

	// Labels 是附加到容器镜像的键值标签。
	Labels map[string]string
}

// BuildConfig 表示 rootfs 打包配置。
type BuildConfig struct {
	// Excludes 是打包 rootfs 时排除的路径列表。
	// 为空时不排除任何路径；如需默认排除路径，请显式使用 DefaultExcludes。
	Excludes []string
}

// Config 表示系统镜像推送配置。
type Config struct {
	// RegistryConfig 是镜像仓库配置。
	RegistryConfig RegistryConfig

	// BuildConfig 是 rootfs 打包配置。
	BuildConfig BuildConfig

	// ImageConfig 是容器镜像配置。
	ImageConfig ImageConfig
}

// DefaultExcludes 返回默认 rootfs 排除路径副本。
func DefaultExcludes() []string {
	return pusher.DefaultExcludes()
}

// PackRootfs 将当前系统 rootfs 打包为 tar 流并返回 reader。
//
// 参数 ctx 用于传递日志上下文。
// 参数 cfg 用于指定 rootfs 打包配置，Excludes 为空时不排除任何路径。
func PackRootfs(ctx context.Context, cfg BuildConfig) io.ReadCloser {
	return pusher.PackRootfs(ctx, pusher.BuildConfig{
		Excludes: cfg.Excludes,
	})
}

// Push 将当前系统 rootfs 打包为容器镜像并推送到指定镜像仓库。
//
// 参数 ctx 用于传递日志上下文。
// 参数 cfg 用于指定镜像仓库、rootfs 打包和容器镜像配置。
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
