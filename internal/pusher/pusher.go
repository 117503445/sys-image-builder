package pusher

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/rs/zerolog/log"
)

var defaultExcludes = []string{
	"/proc", "/sys", "/dev", "/run",
	"/tmp", "/mnt", "/media", "/lost+found",
	"/var/cache", "/var/tmp",
}

type Config struct {
	ImageRef string
	Username string
	Password string
	Insecure bool
	Excludes []string
}

func Push(cfg Config) error {
	excludes := cfg.Excludes
	if len(excludes) == 0 {
		excludes = defaultExcludes
	}

	var opts []name.Option
	if cfg.Insecure {
		opts = append(opts, name.Insecure)
	}

	ref, err := name.ParseReference(cfg.ImageRef, opts...)
	if err != nil {
		return fmt.Errorf("parse image reference: %w", err)
	}

	base, err := mutate.ConfigFile(empty.Image, &v1.ConfigFile{
		Architecture: runtime.GOARCH,
		OS:           "linux",
		Config: v1.Config{
			Cmd: []string{"/bin/sh"},
		},
		RootFS: v1.RootFS{
			Type: "layers",
		},
	})
	if err != nil {
		return fmt.Errorf("set base config: %w", err)
	}

	layer := stream.NewLayer(createRootfsTarReader(excludes))

	img, err := mutate.AppendLayers(base, layer)
	if err != nil {
		return fmt.Errorf("append layer: %w", err)
	}

	auth := &authn.Basic{Username: cfg.Username, Password: cfg.Password}

	log.Info().Str("ref", ref.String()).Msg("pushing image")

	if err := remote.Write(ref, img, remote.WithAuth(auth)); err != nil {
		return fmt.Errorf("push image: %w", err)
	}

	log.Info().Str("ref", ref.String()).Msg("image pushed successfully")
	return nil
}

func shouldExclude(path string, excludes []string) bool {
	for _, exc := range excludes {
		if path == exc || strings.HasPrefix(path, exc+"/") {
			return true
		}
	}
	return false
}

func createRootfsTarReader(excludes []string) io.ReadCloser {
	pr, pw := io.Pipe()

	go func() {
		tw := tar.NewWriter(pw)
		var walkErr error

		walkErr = filepath.Walk("/", func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				if path != "/" {
					log.Warn().Err(err).Str("path", path).Msg("skipping inaccessible path")
					if info != nil && info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
				return err
			}

			if shouldExclude(path, excludes) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			mode := info.Mode()

			// skip sockets and named pipes
			if mode&os.ModeSocket != 0 || mode&os.ModeNamedPipe != 0 {
				return nil
			}

			var linkTarget string
			if mode&os.ModeSymlink != 0 {
				linkTarget, err = os.Readlink(path)
				if err != nil {
					log.Warn().Err(err).Str("path", path).Msg("skipping unreadable symlink")
					return nil
				}
			}

			// use "." prefix for tar paths (like the original package.sh)
			tarPath := "." + path
			if path == "/" {
				tarPath = "./"
			}

			header, err := tar.FileInfoHeader(info, linkTarget)
			if err != nil {
				log.Warn().Err(err).Str("path", path).Msg("skipping file with bad header")
				return nil
			}
			header.Name = tarPath

			// preserve numeric owner IDs
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				header.Uid = int(stat.Uid)
				header.Gid = int(stat.Gid)
			}

			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("write tar header for %s: %w", path, err)
			}

			if mode.IsRegular() {
				f, err := os.Open(path)
				if err != nil {
					log.Warn().Err(err).Str("path", path).Msg("skipping unreadable file")
					return nil
				}
				if _, err := io.Copy(tw, f); err != nil {
					f.Close()
					return fmt.Errorf("write tar body for %s: %w", path, err)
				}
				f.Close()
			}

			return nil
		})

		closeErr := tw.Close()
		if walkErr != nil {
			pw.CloseWithError(walkErr)
		} else if closeErr != nil {
			pw.CloseWithError(closeErr)
		} else {
			pw.Close()
		}
	}()

	return pr
}
