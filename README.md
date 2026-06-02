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
- CPU、内存、磁盘占用展示
- 节点国家、SSH 用户、SSH 目标地址上报
- Reality 域名库与 TLS 延迟探测逻辑
- 一行安装脚本与 systemd 服务模板
- GitHub Actions Release 构建流程

## 一行安装

### 安装主控面板

```bash
bash <(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) master
```

默认建议将主控域名解析到面板服务器：

```text
xui.ityc.cc
```

### 安装 Agent 节点

```bash
bash <(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) agent --master http://xui.ityc.cc:8080 --country China --token test
```

常用参数：

- `--master`：主控面板地址，默认可用 `http://xui.ityc.cc:8080`
- `--token`：节点注册 token，后续会接入正式注册校验
- `--node-id`：自定义节点 ID，默认取系统 hostname
- `--node-name`：面板显示名称，默认取 `--node-id`
- `--country`：节点国家，例如 `China`、`Japan`、`Singapore`、`United States`
- `--endpoint`：SSH 连接目标，默认取节点 hostname，也可填公网 IP 或域名
- `--ssh-user`：SSH 用户，默认 `root`

示例：

```bash
bash <(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) agent --master http://xui.ityc.cc:8080 --node-name hk-01 --country China --endpoint 1.2.3.4 --ssh-user root --token test
```

## Release

正式安装建议先创建一个新 tag，例如：

```bash
git tag v0.1.2
git push origin v0.1.2
```

GitHub Actions 会生成：

- `xui-pro-linux-amd64.tar.gz`
- `xui-pro-linux-arm64.tar.gz`

然后可指定版本安装：

```bash
XUI_PRO_VERSION=v0.1.2 bash <(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) master
```

## 管理命令

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

```bash
go test ./...
go run ./cmd/xuid
```

```bash
XUI_MASTER=http://127.0.0.1:8080 XUI_NODE_ID=hk-01 XUI_NODE_COUNTRY=China go run ./cmd/xui-agent
```

前端开发：

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
5. Reality 域名库定期更新、分国家探测与推荐。
6. 将完整 Vue 3 面板嵌入 `xuid`。
