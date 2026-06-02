# XUI Pro

XUI Pro 是基于 [`sing-web/x-ui`](https://github.com/sing-web/x-ui) 思路重构的现代化 Xray 管理面板。

原项目更适合单台 VPS 使用：Go 后端、Gin、SQLite、嵌入式 Vue 2 模板都在同一个进程里。XUI Pro 保留 Go 的低资源占用优势，但把产品拆成更适合多 VPS 管控的结构：

- `xuid`：主控面板 / Master API
- `xui-agent`：每台 VPS 上运行的轻量 Agent
- `web`：Vue 3 SPA 前端
- `reality`：Reality 域名库与探测推荐模块

## 当前状态

这是一个可继续开发和试装的早期版本，已经具备：

- Master / Agent 基础通信接口
- 节点注册、心跳、基础指标上报
- Reality 域名库与 TLS 延迟探测逻辑
- Vue 3 暗色仪表盘
- 一行安装脚本与 systemd 服务模板
- GitHub Actions Release 构建流程

本地已验证：

```bash
go test ./...
go build ./cmd/xuid
go build ./cmd/xui-agent
cd web && npm run build
```

## 先做 Release

一行安装脚本会从 GitHub Release 下载二进制包，所以首次试装前需要先创建一个版本 tag，例如：

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions 会自动生成：

- `xui-pro-linux-amd64.tar.gz`
- `xui-pro-linux-arm64.tar.gz`

等 Release 资产生成后，再到 VPS 上执行下面的一行安装命令。

## 一行安装

### 安装主控面板

在作为主控的 VPS 上执行：

```bash
bash <(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) master
```

指定监听地址：

```bash
bash <(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) master --listen :8080
```

### 安装 Agent 节点

在其他 VPS 上执行：

```bash
bash <(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) agent --master https://panel.example.com --token xxx --region asia
```

常用参数：

- `--master`：主控面板地址
- `--token`：节点注册 token，后续会接入正式注册校验
- `--node-id`：自定义节点 ID
- `--region`：节点地区，例如 `asia`、`north-america`

## 管理命令

安装后会写入 `xui-pro` 命令：

```bash
xui-pro status
xui-pro restart
xui-pro logs
```

Agent 管理：

```bash
xui-pro agent-status
xui-pro agent-restart
xui-pro agent-logs
```

## 本地开发

启动主控：

```bash
go run ./cmd/xuid
```

启动 Agent：

```bash
XUI_MASTER=http://127.0.0.1:8080 XUI_NODE_ID=hk-01 go run ./cmd/xui-agent
```

启动前端开发服务器：

```bash
cd web
npm install
npm run dev
```

## 路线图

1. SQLite 持久化与数据库迁移。
2. 登录、JWT/OIDC、节点注册 token。
3. Master 到 Agent 的签名配置下发。
4. 迁移并增强原 x-ui 的 Xray inbound/outbound 生成逻辑。
5. Reality 域名库定期更新、分地区探测与推荐。
6. 将 Vue 3 构建产物嵌入 `xuid`，形成完整单二进制面板。
