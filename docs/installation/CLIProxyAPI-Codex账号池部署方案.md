# CLIProxyAPI Codex 账号池部署方案

本文面向已有多个 Codex OAuth 账号、希望通过 `new-api` 统一转换为 API、管理用户和计费并对外分发的部署场景。方案以现有 `new-api` 部署方式为基础，把 Codex OAuth 账号池能力放在 `CLIProxyAPI`，把用户、Token、分组、渠道和计费放在 `new-api`。

推荐架构：

```text
客户端 -> new-api -> CLIProxyAPI -> 多个 Codex OAuth 账号/上游
```

第一阶段不要把 `CLIProxyAPI` 代码直接合并进 `new-api`。两者的职责边界已经清楚：`new-api` 是面向用户、计费和渠道治理的网关，`CLIProxyAPI` 是面向 CLI/OAuth 账号池和协议适配的代理。合并会把 OAuth 凭据刷新、账号冷却、失败切换、管理 API、管理面板发布和 `new-api` 渠道计费逻辑耦合在一起，升级与回滚都会变得更难。第一阶段保持独立服务，通过 OpenAI-compatible 接口集成，风险最低，也便于以后横向扩容多个账号池。

## 职责分工

| 组件 | 职责 | 不负责 |
| --- | --- | --- |
| `new-api` | 用户、Token、分组、模型权限、渠道、失败重试、计费、日志、控制台、对外 API 域名 | Codex OAuth 账号导入、账号刷新、单账号冷却、Codex 协议细节 |
| `CLIProxyAPI` | Codex OAuth 账号池、账号轮询、冷却、失败切换、Codex/OpenAI Responses 协议适配、`/v1` OpenAI-compatible 入口、`/backend-api/codex` Codex 直连别名、`/v0/management` 管理接口 | 用户体系、面向最终用户的计费、`new-api` Token 生命周期 |
| `CPAMC` | 通过 `CLIProxyAPI` Management API 管理 `CLIProxyAPI` 配置、凭据和日志；服务端通常由 `CLIProxyAPI` 在 `/management.html` 提供 | 不参与流量转发，不管理 `new-api` 用户、分组、计费或渠道 |

`new-api` 后台应把 `CLIProxyAPI` 当作一个 OpenAI-compatible 上游渠道，而不是把 Codex OAuth 凭据直接放进 `new-api` 自带 Codex 渠道作为主方案。自带 Codex 渠道适合单独接入或兼容场景；本方案的核心价值是让 `CLIProxyAPI` 集中处理多账号池、冷却、OAuth 刷新和协议适配。

## 架构选择

### 单机 Compose MVP

适合先上线一个 Codex 账号池，所有服务运行在同一台机器、同一个 Compose 网络中：

```text
公网 HTTPS
  |
反向代理
  |
new-api:3000
  |
Compose 内部网络
  |
cli-proxy-api:8317
  |
Codex OAuth 账号池
```

MVP 中只公开 `new-api` 的 Web/API 入口。`CLIProxyAPI` 的 `8317` 端口只在 Compose 网络内给 `new-api` 调用；管理时优先通过服务器本机、临时隧道或受控管理入口访问，不把 `/management.html` 和 `/v0/management` 直接暴露给普通公网。

### 多池可扩展方案

当账号规模、用户分层或风控策略变复杂时，把 `CLIProxyAPI` 拆成多个池：

```text
客户端
  |
new-api
  |-- 渠道 A -> cli-proxy-codex-standard:8317 -> 普通 Codex 账号池
  |-- 渠道 B -> cli-proxy-codex-vip:8317      -> VIP Codex 账号池
  |-- 渠道 C -> cli-proxy-codex-test:8317     -> 测试/灰度账号池
```

推荐拆分维度：

| 拆分方式 | 用法 |
| --- | --- |
| 普通池 / VIP 池 | `new-api` 中配置不同渠道、不同分组和不同定价，VIP 用户只允许访问 VIP 渠道 |
| 稳定池 / 灰度池 | 新模型别名、请求参数或账号策略先在灰度池验证 |
| 地域或网络池 | 每个池使用独立代理、出口或账号来源，降低单点网络故障影响 |
| 团队池 | 每个团队一个 `CLIProxyAPI` 实例，`new-api` 统一用户和计费 |

多池下仍建议让 `new-api` 管渠道级容错和计费，让每个 `CLIProxyAPI` 管池内账号级容错。不要让两个层同时无限重试，否则会放大故障时的请求量。

## 目录结构

建议把运行文件按职责拆开，所有宿主机路径使用相对目录，便于迁移和备份：

```text
codex-pool-stack/
  compose.yml
  new-api/
    data/
    logs/
  cli-proxy/
    config.yaml
    auths/
    logs/
  postgres/
  redis/
```

说明：

| 路径 | 内容 |
| --- | --- |
| `compose.yml` | 本方案的 Compose 文件 |
| `new-api/data/` | `new-api` 容器内 `/data` 的持久化目录 |
| `new-api/logs/` | `new-api` 日志目录 |
| `cli-proxy/config.yaml` | `CLIProxyAPI` 配置文件 |
| `cli-proxy/auths/` | Codex OAuth 凭据目录，由管理面板或 Management API 导入和维护 |
| `cli-proxy/logs/` | `CLIProxyAPI` 日志目录 |
| `postgres/`、`redis/` | 可选的数据库/缓存持久化目录；也可使用命名卷 |

不要把 OAuth 凭据、真实 API key、管理密钥或数据库口令提交到 Git 仓库。

## 密钥与安全边界

至少准备这些密钥，全部使用随机值：

| 密钥 | 用途 | 放置位置 |
| --- | --- | --- |
| `<new-api-session-secret>` | `new-api` 会话密钥，多实例必须一致 | Compose 环境变量或安全的环境文件 |
| `<new-api-crypto-secret>` | `new-api` 加密密钥，启用 Redis 时建议显式设置 | Compose 环境变量或安全的环境文件 |
| `<postgres-password>` | PostgreSQL 口令 | Compose 环境变量 |
| `<redis-password>` | Redis 口令 | Compose 命令和 `REDIS_CONN_STRING` |
| `<cli-proxy-api-key-new-api>` | `new-api` 调用 `CLIProxyAPI /v1` 的内部鉴权 key | `CLIProxyAPI config.yaml` 与 `new-api` 渠道密钥 |
| `<cli-proxy-management-key>` | `CPAMC` / Management API 管理密钥 | `CLIProxyAPI config.yaml` |

生成随机值可以使用：

```bash
openssl rand -hex 32
```

操作原则：

- `new-api` 用户 Token 只发给客户端；`CLIProxyAPI api-keys` 只给 `new-api` 渠道使用。
- `CLIProxyAPI remote-management.secret-key` 只给管理员和 CPAMC 使用，不得填入 `new-api` 渠道。
- Codex OAuth 凭据只通过 CPAMC 或 `/v0/management` 导入，文档、Compose 和 `new-api` 渠道里都不写 OAuth 凭据内容。
- PostgreSQL、Redis、`CLIProxyAPI` 账号池端口不默认对公网开放。
- 公开入口优先只有反向代理后的 `new-api` HTTPS 域名。

## 单机 Compose MVP

以下示例可以用 `podman compose` 或 `docker compose` 执行。二选一即可，后文用 `<compose>` 表示你的 Compose 命令。

```bash
<compose> up -d
<compose> ps
```

### compose.yml 示例

示例中的所有口令均为占位符，部署前必须替换为随机值。

```yaml
services:
  new-api:
    image: calciumion/new-api:latest
    container_name: new-api
    restart: always
    command: --log-dir /app/logs
    ports:
      - "3000:3000"
    volumes:
      - ./new-api/data:/data
      - ./new-api/logs:/app/logs
    environment:
      - TZ=Asia/Shanghai
      - SQL_DSN=postgresql://newapi:<postgres-password>@postgres:5432/newapi
      - REDIS_CONN_STRING=redis://:<redis-password>@redis:6379
      - SESSION_SECRET=<new-api-session-secret>
      - CRYPTO_SECRET=<new-api-crypto-secret>
      - ERROR_LOG_ENABLED=true
      - BATCH_UPDATE_ENABLED=true
      - NODE_NAME=new-api-node-1
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      cli-proxy-api:
        condition: service_healthy
    networks:
      - codex-pool
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O - http://localhost:3000/api/status | grep -o '\"success\":\\s*true' || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3

  cli-proxy-api:
    image: eceasy/cli-proxy-api:latest
    container_name: cli-proxy-api
    restart: unless-stopped
    expose:
      - "8317"
    volumes:
      - ./cli-proxy/config.yaml:/CLIProxyAPI/config.yaml
      - ./cli-proxy/auths:/root/.cli-proxy-api
      - ./cli-proxy/logs:/CLIProxyAPI/logs
    networks:
      - codex-pool
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O - http://localhost:8317/healthz | grep -o 'ok' || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:15
    container_name: new-api-postgres
    restart: always
    environment:
      POSTGRES_USER: newapi
      POSTGRES_PASSWORD: <postgres-password>
      POSTGRES_DB: newapi
    volumes:
      - pg_data:/var/lib/postgresql/data
    networks:
      - codex-pool
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U newapi -d newapi"]
      interval: 30s
      timeout: 10s
      retries: 5

  redis:
    image: redis:latest
    container_name: new-api-redis
    restart: always
    command: ["redis-server", "--requirepass", "<redis-password>", "--appendonly", "yes"]
    volumes:
      - redis_data:/data
    networks:
      - codex-pool
    healthcheck:
      test: ["CMD-SHELL", "redis-cli -a '<redis-password>' ping | grep -q PONG"]
      interval: 30s
      timeout: 10s
      retries: 5

volumes:
  pg_data:
  redis_data:

networks:
  codex-pool:
    driver: bridge
```

如果生产环境需要访问 `CLIProxyAPI` 管理面，优先使用临时隧道、堡垒机、VPN 或仅管理员可访问的反向代理入口。不要在 MVP 示例里默认发布 `8317` 到公网。

### CPAMC 管理入口

MVP 的 `cli-proxy-api` 只使用 `expose: "8317"`，因此宿主机和公网默认不能直接打开 CPAMC。这是安全边界，不是漏配。需要管理账号池时，按场景临时选择一种入口：

| 场景 | 推荐做法 | `remote-management.allow-remote` |
| --- | --- | --- |
| 服务器本机临时管理 | 增加仅本机端口映射，管理完删除映射或停用覆盖文件 | 通常需要临时设为 `true` |
| 运维电脑远程管理 | 先映射到服务器本机，再通过 SSH 隧道连入 | 通常需要临时设为 `true` |
| 长期多人管理 | 放在 VPN、堡垒机或只允许管理员访问的 HTTPS 管理域名后，并加访问控制 | 需要设为 `true` |

原因是 `CLIProxyAPI` 的 Management API 在 `allow-remote: false` 时只接受它自己看到的 localhost 请求。容器通过宿主机端口映射、SSH 隧道或反向代理收到的浏览器请求，通常会被识别为非 localhost；即使只绑定到宿主机 `127.0.0.1`，也可能需要临时开启 `allow-remote`。开启时必须同时满足：管理 key 已设置、入口不对公网开放、只在管理窗口使用或放在受控网络后。

临时本机端口映射可以放到单独覆盖文件，不写入 MVP 主 Compose：

```yaml
# compose.management.yml
services:
  cli-proxy-api:
    ports:
      - "127.0.0.1:8317:8317"
```

启动管理入口：

```bash
<compose> -f compose.yml -f compose.management.yml up -d cli-proxy-api
```

如果人在服务器本机，打开：

```text
http://127.0.0.1:8317/management.html
```

如果从运维电脑管理，先建立 SSH 隧道：

```bash
ssh -N -L 8317:127.0.0.1:8317 <server>
```

然后在运维电脑打开同一个本机地址。管理完成后，把 `remote-management.allow-remote` 改回 `false`，并重新加载或重启 `cli-proxy-api`；同时停用 `compose.management.yml`，让 `8317` 回到仅 Compose 内部可达。

## CLIProxyAPI config.yaml 示例

下面只展示账号池集成所需的最小配置。`api-keys` 是给 `new-api` 调用 `/v1` 的内部 key；`remote-management.secret-key` 是 CPAMC 和 Management API 的管理 key。两者不能混用。

```yaml
host: ""
port: 8317

remote-management:
  allow-remote: false
  secret-key: "<cli-proxy-management-key>"
  disable-control-panel: false

auth-dir: "~/.cli-proxy-api"

api-keys:
  - "<cli-proxy-api-key-new-api>"

debug: false
logging-to-file: true
logs-max-total-size-mb: 1024
error-logs-max-files: 20

request-retry: 3
max-retry-credentials: 3
max-retry-interval: 30
disable-cooling: false

routing:
  strategy: "round-robin"
  session-affinity: true
  session-affinity-ttl: "2h"

quota-exceeded:
  switch-project: true
  switch-preview-model: true
  antigravity-credits: true

oauth-model-alias:
  codex:
    - name: "gpt-5.3-codex"
      alias: "cpa-codex"
```

配置说明：

| 配置 | 建议 |
| --- | --- |
| `api-keys` | 至少配置一个只给 `new-api` 使用的内部 key；客户端不直接使用它 |
| `remote-management.secret-key` | 配置后才启用 `/v0/management`；管理端所有请求都需要这个 key；明文值启动后会被哈希写回 `config.yaml`，因此挂载的配置文件应是可写文件而不是目录 |
| `remote-management.allow-remote` | 业务运行期保持 `false`；用浏览器访问 CPAMC 的临时端口映射、SSH 隧道、VPN 或反向代理入口通常需要设为 `true`，但入口必须限制在本机或受控管理网络内 |
| `routing.strategy` | 第一阶段推荐 `round-robin`，让账号池均衡消耗 |
| `routing.session-affinity` | Codex 会话或长任务建议启用，失败时仍可自动切换可用账号 |
| `request-retry` | 账号级失败由 `CLIProxyAPI` 优先重试，避免把单账号失败直接暴露给 `new-api` |
| `oauth-model-alias.codex` | `name` 是 Codex/OAuth 原始模型名，`alias` 是 `CLIProxyAPI` 对客户端可见的稳定别名 |

如果你使用的是上游 Codex-compatible API key，而不是 OAuth 账号池，可以另外配置 `codex-api-key`。本方案的主路径是 OAuth 账号池，因此不要把 OAuth 凭据写进 `codex-api-key` 示例或文档；用 CPAMC 的“认证文件”或 OAuth 页面导入。

## 导入 Codex OAuth 账号

推荐使用 CPAMC：

1. 启动 `CLIProxyAPI`。
2. 通过受控管理入口打开 `CLIProxyAPI` 的 `/management.html`。
3. 输入 `<cli-proxy-management-key>` 连接。
4. 在“OAuth”或“认证文件”相关页面导入 Codex OAuth 账号。
5. 在“认证文件”页面确认账号启用状态、可见模型、备注和分组信息。
6. 在“系统信息”或模型列表功能中确认 `/v1/models` 能看到 `cpa-codex` 等稳定别名。

也可以使用 `/v0/management` 管理 API 自动化导入。无论使用哪种方式，都不要把 OAuth 凭据正文写入部署文档、工单、聊天记录或 `new-api` 渠道密钥。

## new-api 后台渠道配置

在 `new-api` 控制台创建指向 `CLIProxyAPI` 的 OpenAI-compatible 渠道。

推荐步骤：

1. 进入 `new-api` 管理后台。
2. 打开“渠道”并新建渠道。
3. 渠道类型选择 `OpenAI` / `OpenAI Compatible`。不同前端版本显示可能略有差异，本质是使用 OpenAI-compatible 适配器。
4. 渠道名称填写可识别名称，例如 `Codex Pool - Standard`。
5. API 地址填写：

   ```text
   http://cli-proxy-api:8317
   ```

   不要在这里追加 `/v1`。`new-api` 的 OpenAI 适配器会在调用时拼接 `/v1/chat/completions`、`/v1/responses` 等路径。

6. 渠道密钥填写 `CLIProxyAPI config.yaml` 中专门给 `new-api` 的 `<cli-proxy-api-key-new-api>`。
7. 渠道模型填写 `CLIProxyAPI` 对外暴露的稳定模型名，例如：

   ```text
   cpa-codex
   ```

8. 给渠道绑定允许访问的用户分组，例如 `default`、`vip` 或自定义分组。
9. 配置模型倍率、补全倍率或固定价格，让 `new-api` 成为最终用户计费归属。
10. 保存后执行渠道测试，并再用真实 `new-api` Token 走 `/v1/responses` 或 `/v1/chat/completions` 验证。

### 模型命名与 model_mapping

第一阶段推荐让 `CLIProxyAPI` 输出稳定别名，`new-api` 只做同名暴露：

```text
Codex 原始模型 gpt-5.3-codex
  -> CLIProxyAPI oauth-model-alias 暴露为 cpa-codex
  -> new-api 渠道模型填写 cpa-codex
  -> 客户端请求 cpa-codex
```

这样 `new-api` 不需要 `model_mapping`，计费配置也更直观。

只有当你希望 `new-api` 的客户端可见模型名与 `CLIProxyAPI` 接受的上游模型名不同，才在 `new-api` 渠道中配置 `model_mapping`。`new-api` 的 `model_mapping` 方向是“客户端可见模型名 -> 上游模型名”，例如：

```json
{"cpa-codex":"gpt-5.3-codex"}
```

含义是：客户端请求 `cpa-codex`，`new-api` 转发给上游渠道时改成 `gpt-5.3-codex`。在本方案中，这里的“上游”就是 `CLIProxyAPI`。如果 `CLIProxyAPI` 已经暴露 `cpa-codex`，就不要再配置这条映射，避免后续排障时混淆模型名。

### 不使用 new-api 自带 Codex 渠道作为主方案

本方案不推荐把 `new-api` 自带 Codex 渠道作为第一阶段主路径，原因是：

- 多个 Codex OAuth 账号的刷新、冷却、失败切换和会话粘性已经由 `CLIProxyAPI` 专门处理。
- CPAMC 只能管理 `CLIProxyAPI`，不能管理 `new-api` 内部 Codex 渠道。
- 把 OAuth 凭据直接放进 `new-api` 渠道会削弱账号池集中治理，也会让备份、轮换和审计分散。
- `new-api` 更适合作为最终用户侧网关，统一 Token、分组、日志和计费。

## 网络端口与反向代理

| 服务 | 容器端口 | 默认公开策略 | 说明 |
| --- | --- | --- | --- |
| `new-api` | `3000` | 经反向代理公开 | 对外 Web、管理后台和 OpenAI-compatible API |
| `CLIProxyAPI` | `8317` | 默认仅 Compose 内部可达 | `new-api` 调用 `/v1`，管理员访问 `/management.html` 和 `/v0/management` |
| PostgreSQL | `5432` | 不公开 | 只给 `new-api` 使用 |
| Redis | `6379` | 不公开 | 只给 `new-api` 使用 |
| `CLIProxyAPI` pprof | `8316` | 不启用或仅本机 | 仅排障临时使用 |

反向代理原则：

- 公共用户只访问 `new-api`，例如 `https://<new-api-domain>/v1/...`。
- 代理必须支持流式响应，避免对 SSE 或长连接启用破坏流式传输的缓冲。
- `new-api` 到 `CLIProxyAPI` 使用 Compose 服务名 `http://cli-proxy-api:8317`，不绕出公网。
- 管理端 `/management.html` 和 `/v0/management` 不与用户 API 共用公开入口；确需远程访问时使用独立管理域名、VPN、访问控制和 HTTPS。
- 数据库、Redis 和 OAuth 凭据目录只在服务器内部或备份系统中可见。

## 验证流程

### 1. 验证 CLIProxyAPI 健康状态

在 Compose 网络内执行：

```bash
<compose> exec cli-proxy-api wget -q -O - http://localhost:8317/healthz
```

期望返回：

```json
{"status":"ok"}
```

### 2. 验证 CLIProxyAPI 模型列表

```bash
<compose> exec cli-proxy-api wget -q -O - \
  --header "Authorization: Bearer <cli-proxy-api-key-new-api>" \
  http://localhost:8317/v1/models
```

期望结果中能看到 `cpa-codex` 或你配置的稳定别名。若返回 401，检查 `api-keys` 是否与 `new-api` 渠道密钥一致。

### 3. 验证 new-api 健康状态

```bash
curl -sS http://localhost:3000/api/status
```

期望 JSON 中 `success` 为 `true`。

### 4. 验证 new-api Token 调用 Responses

使用 `new-api` 后台创建的用户 Token：

```bash
curl -sS http://localhost:3000/v1/responses \
  -H "Authorization: Bearer <new-api-user-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "cpa-codex",
    "input": "请返回 pong"
  }'
```

如果客户端还只支持 Chat Completions，可以验证：

```bash
curl -sS http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer <new-api-user-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "cpa-codex",
    "messages": [
      {"role": "user", "content": "请返回 pong"}
    ]
  }'
```

验证通过后，再检查 `new-api` 日志中是否产生正确的用户、Token、模型、渠道和计费记录。

## 多池落地方式

### Compose 内多池

同一台机器上可以复制多个 `CLIProxyAPI` 服务，每个池使用自己的 `config.yaml`、`auths/` 和日志目录：

```yaml
services:
  cli-proxy-codex-standard:
    image: eceasy/cli-proxy-api:latest
    restart: unless-stopped
    expose:
      - "8317"
    volumes:
      - ./pools/standard/config.yaml:/CLIProxyAPI/config.yaml
      - ./pools/standard/auths:/root/.cli-proxy-api
      - ./pools/standard/logs:/CLIProxyAPI/logs
    networks:
      - codex-pool

  cli-proxy-codex-vip:
    image: eceasy/cli-proxy-api:latest
    restart: unless-stopped
    expose:
      - "8317"
    volumes:
      - ./pools/vip/config.yaml:/CLIProxyAPI/config.yaml
      - ./pools/vip/auths:/root/.cli-proxy-api
      - ./pools/vip/logs:/CLIProxyAPI/logs
    networks:
      - codex-pool
```

然后在 `new-api` 中创建两个渠道：

| new-api 渠道 | API 地址 | 分组 | 定价建议 |
| --- | --- | --- | --- |
| `Codex Pool - Standard` | `http://cli-proxy-codex-standard:8317` | `default` | 普通价格 |
| `Codex Pool - VIP` | `http://cli-proxy-codex-vip:8317` | `vip` | VIP 价格或更高倍率 |

### 多机多池

多机时建议每台机器运行一个或多个 `CLIProxyAPI` 池，`new-api` 通过内网或受控网络访问。每个池仍是一个独立渠道，避免在 `new-api` 之外再套一层不透明负载均衡，除非你能把健康检查、故障摘除和日志关联一起做好。

## 上线前决策审问清单

| 问题 | 推荐答案 |
| --- | --- |
| 第一阶段是单池还是多池？ | 先单池 MVP，上线后按 VIP、团队、地域或灰度需求拆多池。 |
| 最终用户计费归属在哪里？ | 放在 `new-api`。`CLIProxyAPI` 的使用量和日志只做账号池排障与容量观察。 |
| 是否把 `CLIProxyAPI` 管理端暴露到公网？ | 默认不暴露。必须远程管理时，使用独立管理入口、HTTPS、访问控制和强管理密钥。 |
| CPAMC 管理入口怎么开？ | 默认不开宿主机端口；临时管理用本机端口映射加 SSH 隧道，长期管理用 VPN、堡垒机或独立受控管理域名。 |
| `remote-management.allow-remote` 何时开启？ | 业务运行期保持 `false`；浏览器 CPAMC 经过端口映射、隧道或反向代理访问时才临时或受控开启。 |
| 是否拆 VIP 池？ | 有明确 SLA、不同价格或高价值账号时拆；否则先用一个池降低复杂度。 |
| 重试归谁管？ | 账号级失败由 `CLIProxyAPI` 管，渠道级失败由 `new-api` 管；两层都要设上限。 |
| 模型名谁定义？ | 第一阶段由 `CLIProxyAPI` 用 `oauth-model-alias` 输出稳定别名，`new-api` 同名暴露。 |
| 什么时候用 `new-api model_mapping`？ | 只有公开名与 `CLIProxyAPI` 可接受模型名需要不同才用；方向是客户端可见名映射到上游名。 |
| 是否启用会话粘性？ | Codex 长会话和多轮任务建议在 `CLIProxyAPI` 启用 `routing.session-affinity`。 |
| OAuth 凭据怎么备份？ | 备份 `cli-proxy/auths/` 和 `config.yaml` 到加密介质；恢复前确认权限和管理 key。 |
| 是否使用 `new-api` 自带 Codex 渠道？ | 不作为主路径。保留为特殊兼容或单独测试，不承担多账号池主方案。 |
| 公开给客户端的 Base URL 是什么？ | 只公开 `new-api` 的 `/v1` 入口，客户端不直接调用 `CLIProxyAPI`。 |
| 是否让 CPAMC 管理 `new-api`？ | 不。CPAMC 只管理 `CLIProxyAPI`；`new-api` 用自己的后台管理用户、Token、渠道和计费。 |

## 备份

必须备份：

| 对象 | 备份内容 | 频率建议 |
| --- | --- | --- |
| `new-api` 主数据库 | PostgreSQL 数据库，或 SQLite 文件 | 每日，升级前额外备份 |
| `new-api` 数据目录 | `new-api/data/` | 每日，升级前额外备份 |
| `new-api` 日志 | `new-api/logs/` 或独立日志库 | 按审计要求 |
| `CLIProxyAPI` 配置 | `cli-proxy/config.yaml` | 每次变更后 |
| Codex OAuth 凭据 | `cli-proxy/auths/` | 每次导入、刷新策略变更或账号调整后 |
| Compose 文件 | `compose.yml` | 每次变更后 |

恢复顺序：

1. 停止服务。
2. 恢复数据库和持久化目录。
3. 恢复 `CLIProxyAPI config.yaml` 与 `auths/`。
4. 确认密钥与环境变量匹配。
5. 启动 `CLIProxyAPI`，验证 `/healthz` 和 `/v1/models`。
6. 启动 `new-api`，验证 `/api/status` 和真实 Token 调用。

OAuth 凭据属于高敏数据，备份必须加密并限制访问。不要把备份路径、密钥位置或凭据内容写入公开文档。

## 升级与回滚

### 升级前

1. 记录当前镜像标签或镜像摘要。
2. 备份 `new-api` 数据库、`new-api/data/`、`cli-proxy/config.yaml` 和 `cli-proxy/auths/`。
3. 在 CPAMC 中确认账号池健康，记录当前可用模型列表。
4. 如果使用 `latest`，生产环境建议改为明确标签或摘要，便于回滚。

### 升级

```bash
<compose> pull
<compose> up -d
<compose> ps
<compose> logs -f new-api
<compose> logs -f cli-proxy-api
```

升级后依次执行：

1. `CLIProxyAPI /healthz`
2. `CLIProxyAPI /v1/models`
3. `new-api /api/status`
4. `new-api /v1/responses` 或 `/v1/chat/completions`
5. `new-api` 后台渠道测试

### 回滚

如果升级后失败：

1. 把镜像标签改回升级前版本。
2. 执行 `<compose> up -d`。
3. 如果数据库已发生不兼容迁移，停止服务后恢复升级前数据库备份。
4. 恢复 `CLIProxyAPI config.yaml` 和 `auths/` 备份。
5. 重新执行完整验证流程。

回滚时不要只回滚一个服务后立刻放量；先确认 `new-api` 渠道和 `CLIProxyAPI` 模型列表仍一致。

## 监控与日志

最低监控项：

| 项目 | 观察位置 |
| --- | --- |
| `new-api` 存活 | `/api/status`、Compose healthcheck、反向代理 5xx |
| `CLIProxyAPI` 存活 | `/healthz`、Compose healthcheck |
| 模型列表 | `CLIProxyAPI /v1/models` 与 `new-api` 渠道模型是否一致 |
| 计费记录 | `new-api` 日志、用户额度、模型倍率和渠道命中 |
| 账号池状态 | CPAMC 认证文件、配额、请求日志、错误日志 |
| 流式稳定性 | 客户端超时、反向代理日志、`new-api` 和 `CLIProxyAPI` 错误日志 |
| 数据库/Redis | 容器状态、连接错误、磁盘占用、备份任务结果 |

日志建议：

- `new-api` 开启 `ERROR_LOG_ENABLED=true`，并挂载 `./new-api/logs:/app/logs`。
- `CLIProxyAPI` 开启 `logging-to-file: true` 并挂载 `./cli-proxy/logs:/CLIProxyAPI/logs`。
- 管理端日志和请求日志只给管理员访问，不公开给最终用户。
- 如果需要长期账号池用量统计，优先让 `new-api` 作为最终计费源；`CLIProxyAPI` 侧可按需要接入独立用量看板。

## 排障矩阵

| 现象 | 重点检查 | 处理建议 |
| --- | --- | --- |
| `new-api /api/status` 不通 | `new-api` 容器状态、端口映射、数据库连接、Redis 连接 | 查看 `new-api` 日志；确认 `SQL_DSN`、`REDIS_CONN_STRING` 和密钥占位符已替换 |
| `CLIProxyAPI /healthz` 不通 | `cli-proxy-api` 容器状态、`config.yaml` 挂载、端口 `8317` | 查看 `CLIProxyAPI` 日志；确认配置文件 YAML 格式正确 |
| `/v1/models` 返回 401 | `Authorization` 是否使用 `api-keys`，而不是 management key | 用 `<cli-proxy-api-key-new-api>` 重试；同步更新 `new-api` 渠道密钥 |
| `/v1/models` 没有 Codex 模型 | OAuth 账号是否导入、启用、可用；`oauth-model-alias` 是否写在 `codex` 下 | 在 CPAMC 检查认证文件和模型列表；重新加载配置 |
| `new-api` 渠道测试失败 | API 地址是否写成 `http://cli-proxy-api:8317`，是否误加 `/v1` | 去掉 `/v1`，确认 Compose 服务名和网络一致 |
| 客户端模型不存在 | `new-api` 渠道模型、`CLIProxyAPI` 模型别名、`model_mapping` 方向 | 第一阶段删除 `new-api model_mapping`，改为同名暴露再测 |
| Responses 可用但 Chat Completions 异常 | 客户端接口类型、模型是否适配 Chat Completions、转换路径 | 优先用 `/v1/responses` 验证 Codex；必要时为客户端单独配置兼容模型 |
| 流式中断或空响应 | 反向代理缓冲、超时、`STREAMING_TIMEOUT`、上游账号冷却 | 关闭破坏 SSE 的代理缓冲；调大超时；检查账号池错误日志 |
| 请求重复计费或失败后费用异常 | `new-api` 重试次数、`CLIProxyAPI request-retry`、客户端重试 | 限制各层重试次数；把账号级重试留给 `CLIProxyAPI` |
| CPAMC 无法连接 | `/management.html` 是否可达、`remote-management.secret-key` 是否配置、远程访问是否允许 | 本机管理保持 `allow-remote: false`；远程管理需受控入口并启用远程访问 |
| 升级后模型名变化 | `oauth-model-alias`、Codex 上游模型列表、`new-api` 渠道模型 | 固定 `CLIProxyAPI` 稳定别名；必要时用灰度池先验证 |
| 登录状态不稳定 | `new-api SESSION_SECRET` 是否固定，多实例是否一致 | 设置稳定随机值，不要每次启动生成新值 |
| Redis 连接失败 | Redis 口令、连接串、服务名、网络 | 确认 `REDIS_CONN_STRING=redis://:<redis-password>@redis:6379` 与 Redis 命令一致 |

## 最小上线步骤

1. 准备目录结构和随机密钥。
2. 写入 `compose.yml` 和 `cli-proxy/config.yaml`。
3. 启动 Compose。
4. 初始化 `new-api` 管理员。
5. 通过 CPAMC 导入 Codex OAuth 账号。
6. 验证 `CLIProxyAPI /healthz` 和 `/v1/models`。
7. 在 `new-api` 新建 OpenAI-compatible 渠道，Base URL 指向 `http://cli-proxy-api:8317`。
8. 配置 `new-api` 模型名、分组和价格。
9. 用 `new-api` 用户 Token 验证 `/v1/responses` 或 `/v1/chat/completions`。
10. 配置反向代理，只公开 `new-api`。
11. 建立备份、升级和回滚流程。
12. 小流量放量，观察日志、计费和账号池冷却情况。
