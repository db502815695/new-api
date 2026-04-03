# Caddy Docker Sysctl 运维手册设计

**日期**: 2026-04-04

**目标**: 新增一份面向 `2C4G` 中转机的中文运维手册，统一收录 Caddy 反代配置、Docker 容器调优、`/etc/sysctl.conf` 系统参数调优与验证步骤，便于直接落地执行。

**设计范围**:

- 文档路径固定为 `docs/installation/caddy-docker-sysctl.md`
- 内容聚焦 `Caddy -> Go 中转 -> 上游模型` 链路
- 优先覆盖 SSE、WebSocket、长首包等待场景
- 参数取值以“稳中求性能”为原则，不采用大机器激进配置

**核心决策**:

1. Caddy 配置采用 `/ws` 单独分流，其他请求统一走 API/SSE 路由。
2. Caddy 到上游使用 HTTP/1.1，并关闭上游压缩，避免流式响应被压缩攒包。
3. Docker 文档明确要求显式配置 `ulimits.nofile`，避免仅依赖宿主机继承。
4. 系统调优重点放在 `/etc/sysctl.conf` 中的 `net.core.somaxconn` 和 `net.netfilter.nf_conntrack_max`。
5. `2C4G`、约 `2000` 并发连接目标下，推荐 `somaxconn=16384`、`nf_conntrack_max=131072`。

**验收标准**:

- 文档使用简体中文
- 文档包含可直接复制的 Caddyfile 片段
- 文档包含可直接复制的 Docker Compose 片段
- 文档包含 `/etc/sysctl.conf` 配置示例与生效命令
- 文档包含验证命令和风险说明
