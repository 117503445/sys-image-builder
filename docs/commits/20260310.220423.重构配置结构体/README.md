# 重构配置结构体

## 主要内容和目的

重构 `sysimage.Config` 结构体，将配置按职责分离为 `RegistryConfig`、`BuildConfig` 和 `ImageConfig` 三个独立结构体，提高 API 的可读性和可维护性。

## 更改内容描述

1. **新增 `RegistryConfig` 结构体**
   - 包含 `ImageRef`、`Username`、`Password`、`Insecure` 字段
   - 用于配置 registry 相关参数

2. **新增 `BuildConfig` 结构体**
   - 包含 `Excludes` 字段
   - 用于配置构建相关参数

3. **新增 `ImageConfig` 结构体**
   - 包含 `Cmd`、`Entrypoint`、`WorkingDir`、`User`、`ExposedPorts`、`Env`、`Labels` 字段
   - 用于配置容器镜像元数据

4. **修改 `Config` 结构体**
   - 改为组合 `RegistryConfig`、`BuildConfig`、`ImageConfig`

5. **同步修改 `internal/pusher/pusher.go`**
   - 相应调整配置结构体
   - 添加默认值处理：若 `Cmd` 和 `Entrypoint` 都为空，默认设置 `Cmd: ["/bin/sh"]`

## 验证方法和结果

运行 e2e 测试：`go run ./scripts/go-script e2e`

结果：**E2E TEST PASSED**