package sysimage

import (
	"archive/tar"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPackRootfsReturnsReader 验证 PackRootfs 会返回可读取的 rootfs tar 流。
func TestPackRootfsReturnsReader(t *testing.T) {
	tempDir := t.TempDir()
	tempDir, err := filepath.EvalSymlinks(tempDir)
	if err != nil {
		t.Fatalf("解析临时目录失败: %v", err)
	}

	filePath := filepath.Join(tempDir, "payload.txt")
	if err := os.WriteFile(filePath, []byte("hello-rootfs-reader"), 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	excludes, err := buildExcludesExceptPath(tempDir)
	if err != nil {
		t.Fatalf("构造排除路径失败: %v", err)
	}

	reader := PackRootfs(context.Background(), BuildConfig{Excludes: excludes})
	defer reader.Close()

	wantName := "." + filePath
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("读取 tar 流失败: %v", err)
		}
		if header.Name != wantName {
			continue
		}

		content, err := io.ReadAll(tarReader)
		if err != nil {
			t.Fatalf("读取 tar 文件内容失败: %v", err)
		}
		if string(content) != "hello-rootfs-reader" {
			t.Fatalf("文件内容不匹配，期望 %q，实际 %q", "hello-rootfs-reader", string(content))
		}
		return
	}

	t.Fatalf("未在 rootfs tar 流中找到 %s", wantName)
}

// TestDefaultExcludesReturnsCopy 验证默认排除路径需要显式获取，且返回值可安全修改。
func TestDefaultExcludesReturnsCopy(t *testing.T) {
	excludes := DefaultExcludes()
	if len(excludes) == 0 {
		t.Fatal("默认排除路径不能为空")
	}

	excludes[0] = "/changed"
	if DefaultExcludes()[0] == "/changed" {
		t.Fatal("DefaultExcludes 应返回副本，不能暴露内部切片")
	}
}

// buildExcludesExceptPath 为 rootfs 打包测试构造排除列表，仅保留指定路径。
func buildExcludesExceptPath(path string) ([]string, error) {
	cleanPath := filepath.Clean(path)
	relativePath, err := filepath.Rel("/", cleanPath)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(relativePath, string(os.PathSeparator))
	excludes := make([]string, 0)
	current := string(os.PathSeparator)

	for _, part := range parts {
		keepPath := filepath.Join(current, part)
		entries, err := os.ReadDir(current)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			entryPath := filepath.Join(current, entry.Name())
			if entryPath != keepPath {
				excludes = append(excludes, entryPath)
			}
		}

		current = keepPath
	}

	return excludes, nil
}
