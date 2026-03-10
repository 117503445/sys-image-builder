# 新增 pkg/sysimage 公开 API

## 主要内容和目的

创建 `pkg/sysimage` 包，暴露公开 API 供其他 Go 程序集成，实现将系统 rootfs 打包成容器镜像并推送到 registry 的功能。

## 更改内容描述

1. **新增 `pkg/sysimage/sysimage.go`**
   - 定义 `Config` 结构体，包含镜像引用、认证信息、排除路径等配置
   - 提供 `Push(ctx, cfg)` 方法，支持 context 参数

2. **修改 `internal/pusher/pusher.go`**
   - `Push` 函数新增 `context.Context` 参数
   - 使用 `log.Ctx(ctx)` 打印日志，支持上下文日志

3. **修改 `cmd/cli/push.go`**
   - 改为引用 `pkg/sysimage` 而非 `internal/pusher`
   - 创建带 logger 的 context 传递给 Push 方法

## 验证方法和结果

运行 e2e 测试：`go run ./scripts/go-script e2e`

测试流程：
1. 构建静态二进制
2. 启动带认证的 registry
3. 启动 alpine 容器并创建测试文件
4. 推送 alpine 文件系统到 registry
5. 拉取镜像并验证测试文件内容

结果：**E2E TEST PASSED**