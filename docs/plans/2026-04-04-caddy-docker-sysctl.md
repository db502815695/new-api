# Caddy Docker Sysctl 运维手册 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 新增一份单文件运维手册，统一说明 `2C4G` 中转机场景下的 Caddy、Docker 和 `/etc/sysctl.conf` 调优建议。

**Architecture:** 手册放在 `docs/installation/` 目录，正文按“适用范围 -> Caddy -> Docker -> sysctl -> 验证 -> 回滚”的顺序组织，所有参数取值围绕 SSE、WebSocket 和长首包等待场景展开。计划同步生成设计记录，便于后续追溯决策依据。

**Tech Stack:** Markdown, Git

---

### Task 1: 落设计记录

**Files:**
- Create: `docs/plans/2026-04-04-caddy-docker-sysctl-design.md`

**Step 1: 写入设计文档**

编写设计目标、范围、关键参数取值与验收口径。

**Step 2: 自查设计内容**

检查是否覆盖 Caddy、Docker、`sysctl` 三部分。

**Step 3: 暂存设计文档**

Run: `git add docs/plans/2026-04-04-caddy-docker-sysctl-design.md`

Expected: 文件进入暂存区。

### Task 2: 编写正式运维手册

**Files:**
- Create: `docs/installation/caddy-docker-sysctl.md`

**Step 1: 编写适用场景**

说明机器规格、并发目标和配置边界。

**Step 2: 编写 Caddy 配置**

给出 `/ws` 分流、SSE/API 反代、超时和压缩策略。

**Step 3: 编写 Docker 调优**

给出 `ulimits.nofile`、重启策略和简单运行建议。

**Step 4: 编写 `/etc/sysctl.conf` 调优**

给出推荐值、临时生效命令、永久生效方式和验证命令。

**Step 5: 编写回滚与风险说明**

说明参数过大或误配时的排查方向和回滚方法。

### Task 3: 编写实施计划记录

**Files:**
- Create: `docs/plans/2026-04-04-caddy-docker-sysctl.md`

**Step 1: 编写计划头部**

写明目标、架构和技术栈。

**Step 2: 拆分任务**

按设计记录、正式手册、校对、提交和推送顺序列出任务。

### Task 4: 校对并提交

**Files:**
- Modify: `docs/plans/2026-04-04-caddy-docker-sysctl-design.md`
- Modify: `docs/plans/2026-04-04-caddy-docker-sysctl.md`
- Modify: `docs/installation/caddy-docker-sysctl.md`

**Step 1: 检查 Markdown**

Run: `git diff -- docs/plans/2026-04-04-caddy-docker-sysctl-design.md docs/plans/2026-04-04-caddy-docker-sysctl.md docs/installation/caddy-docker-sysctl.md`

Expected: 仅包含新增文档内容。

**Step 2: 提交**

Run: `git add docs/plans/2026-04-04-caddy-docker-sysctl-design.md docs/plans/2026-04-04-caddy-docker-sysctl.md docs/installation/caddy-docker-sysctl.md && git commit -m "docs: add caddy docker sysctl operations guide"`

Expected: 生成一条文档提交记录。

### Task 5: 推送远端

**Files:**
- None

**Step 1: 推送**

Run: `git push origin main`

Expected: 远端 `origin/main` 更新成功。
