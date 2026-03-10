# feat(cli): 新增 push 命令支持流式推送系统镜像到 registry

## 主要内容和目的

实现将当前系统文件系统打包为 OCI 镜像并推送到远程 registry 的功能。核心特点是采用流式处理，不占用额外的磁盘空间：

- 流式读取文件 -> 添加到 tar.gz 流 -> 流式推送到远程 registry

## 更改内容描述

### 新增文件

1. `cmd/cli/push.go` - push 命令入口
2. `internal/pusher/pusher.go` - 核心推送逻辑实现
   - 使用 `io.Pipe` 实现流式 tar 打包
   - 使用 `go-containerregistry` 库构建和推送 OCI 镜像
   - 支持 Basic Auth 认证
   - 支持 HTTP (insecure) 模式
   - 自动排除 /proc, /sys, /dev 等系统目录
3. `scripts/go-script/` - E2E 测试脚本
   - 编译静态二进制
   - 启动带认证的 registry 容器
   - 启动 alpine 测试容器
   - 在容器内执行 push 命令
   - 验证推送的镜像内容正确
4. `scripts/tasks/e2e/Taskfile.yml` - E2E 测试任务定义

### 修改文件

1. `cmd/cli/cli.go` - 添加 `CmdPush` 命令定义
2. `go.mod` / `go.sum` - 添加 `go-containerregistry` 依赖
3. `.gitignore` - 添加 `p.md` 和 `/cli` 排除规则
4. `Taskfile.yml` - 添加 e2e 任务引用

## 验证方法和结果

通过 E2E 测试验证：

```bash
task e2e:run
```

测试流程：
1. 编译 sys-image-builder 静态二进制
2. 启动带 htpasswd 认证的 registry 容器
3. 启动 alpine 容器并复制二进制
4. 在 alpine 容器内创建 `/root/test.txt` 测试文件
5. 执行 push 命令将容器文件系统推送到 registry
6. 在宿主机拉取镜像并验证 `/root/test.txt` 内容正确

预期结果：E2E TEST PASSED