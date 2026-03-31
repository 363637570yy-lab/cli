# CLIProxyAPI (防风控特制版)

本版本基于原版 [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) 进行了深度定制和优化。主要针对近期 OpenAI 严厉封禁 Codex CLI 凭证的“风控策略”，在服务端链路层面彻底切断了任何可能暴露第三方非官方工具特征的风险点。

---

## 🚀 核心优化与差异点

相比于上游原版，本定制版本主要解决了 **指纹泄露** 和 **请求环境特征不匹配** 导致的高危封禁问题：

### 1. 彻底斩断下游客户端特征透传 (防止非法工具特征泄露)
*   **原版行为**：程序会自动放行和透传下游（如你的本地电脑）发来的 `User-Agent` 或 `Originator` 识别串。如果使用 `claude-code` 或 `curl` 请求该代理，这些特殊的客户端请求头会被如实发往 OpenAI，导致瞬间被风控系统识别拦截。
*   **本版行为**：引入了**强制洗白**机制。我们在 `codex_executor.go` 和 `codex_websockets_executor.go` 核心请求节点处，强行拦截并丢弃了任何来自下游的 `User-Agent` 与 `Originator`，统一将发往 OpenAI 的请求伪装成最纯净的 `codex_cli_rs` 与 `codex_originator`。

### 2. 硬编码安全客户端运行环境 (防止服务端平台被识破)
*   **原版行为**：在发起主请求和 `OAuth` 刷新 Token 时，会利用 Go 语言内置的 `runtime.GOOS` 动态抓取当前运行系统（如 `Linux 6.1.0`）。由于绝大多数真实 Codex 用户运行在 Windows/macOS 桌面端，在 Linux 上进行高频并发请求很容易被判定为异常服务端代理。
*   **本版行为**：剥离了动态系统探测代码。不管把 `CLIProxyAPI` 部署在任何 Linux 服务器上，它的主请求以及后台静默 Token 刷新请求，都会全部**强制硬编码生成基于真实 Windows 环境的 User-Agent**： 
    `(Windows 10.0.22631; x86_64)`
    与你本地的 Windows 操作环境保持上下文绝对一致，打散了原本矛盾的环境特征。

---

## 📦 如何编译与部署该定制版

在你的 Linux 服务器（如 `www.qnzn.top`）上更新和部署此版本非常简单。
每次你往自己的 GitHub 仓库 (`363637570yy-lab/cli`) 提交代码后，请在 SSH 终端执行以下一键部署命令集：

```bash
# 1. 切换到源码拉取目录
cd ~/CLIProxyAPI-build

# 2. 从你自己的私有库同步最新定制代码
git pull origin main

# 3. 配置 Go 环境变量并编译适用于 Linux 的无 C 依赖二进制文件
GOROOT=~/go GOPATH=~/gopath PATH=~/go/bin:$PATH CGO_ENABLED=0 go build -o cli-proxy-api ./cmd/server

# 4. 停止旧服务并替换二进制程序
systemctl --user stop cliproxyapi
sleep 1
cp cli-proxy-api ~/cliproxyapi/cli-proxy-api
chmod +x ~/cliproxyapi/cli-proxy-api

# 5. 重启服务并检查状态
systemctl --user start cliproxyapi
sleep 2
systemctl --user is-active cliproxyapi
```

*(如果终端输出 `active`，即表示本定制版上线成功！)*

---

## 🔄 原版上游如果有大更新，我该怎么办？

由于上游 [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) 可能会发布新功能（比如支持新模型、新协议），你绝不能将服务器重新切回上游原版（一旦切回又会面临秒封风险）。
**正确的做法是：将上游的新功能合并（Merge）到你现在的定制版中。**

你可以按照以下步骤在本地的 `D:\BCKF\codex\CLIProxyAPI` 终端内操作：

### 步骤 1：添加上游原版作为远程仓库 (仅需做一次)
```bash
git remote add upstream https://github.com/router-for-me/CLIProxyAPI.git
```

### 步骤 2：拉取上游的最新更新
```bash
git fetch upstream
```

### 步骤 3：将上游的新功能合并到你的定制版分支中
```bash
git merge upstream/main
```
*(如果在这步遇到了代码冲突：多数时候是因为上游也改了请求头相关的代码。不用慌张，可以通过 VSCode 或 Cursor 手动解决冲突，原则是**始终保留我们硬编码的 Windows 指纹和强转下游拦截逻辑**即可。)*

### 步骤 4：推送更新到你的个人仓库，并在服务器重新编译
```bash
git push mylab main
```
最后去服务器执行上面提到的【一键部署命令集】即可完成升级。
