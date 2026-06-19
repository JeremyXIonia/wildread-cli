# Release

## 发布新版本

确保当前在 `master` 且测试通过：

```bash
go test ./...
```

创建并推送 tag：

```bash
git tag v0.1.0
git push origin v0.1.0
```

推送 tag 后，GitHub Actions 会自动：

- 运行测试
- 构建 macOS Intel、macOS Apple Silicon、Windows amd64 二进制
- 打包 Release artifacts
- 生成 `checksums.txt`
- 创建 GitHub Release

## 版本号

使用 SemVer 风格 tag：

- `v0.1.0`
- `v0.2.0`
- `v1.0.0`

## Release artifacts

每次 release 应包含：

- `wildread-cli-darwin-amd64.tar.gz`
- `wildread-cli-darwin-arm64.tar.gz`
- `wildread-cli-windows-amd64.zip`
- `checksums.txt`
