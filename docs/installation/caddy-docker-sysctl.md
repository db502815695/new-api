# Caddy + Docker + `/etc/sysctl.conf` 运维手册

## 1. 适用场景

这份手册适用于以下场景：

- 机器规格约为 `2C4G`
- 服务链路为 `客户端 -> Caddy -> Go 中转 -> 上游模型`
- 主要业务为 API 中转，包含 SSE 流式输出和 `/ws` WebSocket
- 目标是约 `2000` 并发连接，追求稳中求性能，不追求极限压榨

这份手册不以“大机器极限参数”为目标。所有参数均按小机器可落地、便于维护的原则给出。

## 2. Caddy 配置建议

建议将 WebSocket 和普通 API/SSE 分开处理。`/ws` 单独走一条反代，其他请求走统一的 API/SSE 反代。

推荐配置如下：

```caddyfile
your.domain.com {

	@ws path /ws*

	handle @ws {
		reverse_proxy 127.0.0.1:8080 {
			stream_timeout 24h
			stream_close_delay 5m

			transport http {
				versions 1.1
				compression off
				dial_timeout 5s
				response_header_timeout 90s
				keepalive 2m
				keepalive_idle_conns 256
				keepalive_idle_conns_per_host 128
			}
		}
	}

	handle {
		header {
			Cache-Control "no-store, no-cache, must-revalidate"
			Pragma "no-cache"
			X-Accel-Buffering "no"
		}

		reverse_proxy 127.0.0.1:8080 {
			transport http {
				versions 1.1
				compression off
				dial_timeout 5s
				response_header_timeout 90s
				keepalive 2m
				keepalive_idle_conns 256
				keepalive_idle_conns_per_host 128
			}
		}
	}
}
```

配置说明：

- `@ws path /ws*`
  将 WebSocket 路径单独分流，避免普通请求混入升级逻辑。
- `compression off`
  关闭 Caddy 到上游的压缩协商，减少流式响应被压缩攒包的概率。
- `response_header_timeout 90s`
  适配上游首包可能到 `60s` 的情况，预留一定缓冲。
- `keepalive`
  复用到 Go 中转的连接，减少重复建连开销。
- `stream_close_delay 5m`
  降低 Caddy reload 时 WebSocket 连接的抖动。

不建议默认启用 `flush_interval -1`。这个参数更激进，在客户端提前断开时可能让后端请求继续运行，不符合“稳中求性能”的目标。

## 3. Docker 调优建议

如果 Caddy 或 Go 中转运行在 Docker 容器中，建议显式声明 `nofile`，不要仅依赖宿主机继承。

`docker-compose.yml` 示例：

```yaml
services:
  caddy:
    image: caddy:latest
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    ulimits:
      nofile:
        soft: 262144
        hard: 262144

  new-api:
    image: your/new-api:latest
    restart: unless-stopped
    ulimits:
      nofile:
        soft: 262144
        hard: 262144
```

建议说明：

- `restart: unless-stopped`
  适合生产环境的基础重启策略。
- `ulimits.nofile`
  保证容器内部也能拿到足够的文件句柄上限。
- `262144`
  对 `2C4G` 小机器已足够宽裕，不需要继续抬高。

宿主机检查命令：

```bash
ulimit -n
```

容器内检查命令：

```bash
docker exec -it <container_name> sh -c 'ulimit -n'
```

## 4. `/etc/sysctl.conf` 调优建议

对于 `2C4G`、约 `2000` 并发连接的中转机场景，建议优先调整以下两项：

```conf
net.core.somaxconn = 16384
net.netfilter.nf_conntrack_max = 131072
```

含义说明：

- `net.core.somaxconn`
  控制监听队列上限。连接突发较多时，过小会导致接入层排队能力不足。
- `net.netfilter.nf_conntrack_max`
  控制连接跟踪表上限。Docker、NAT、iptables 场景下很重要，太小容易造成连接异常。

直接写入 `/etc/sysctl.conf` 即可，例如：

```conf
# /etc/sysctl.conf
net.core.somaxconn = 16384
net.netfilter.nf_conntrack_max = 131072
```

临时生效命令：

```bash
sysctl -w net.core.somaxconn=16384
sysctl -w net.netfilter.nf_conntrack_max=131072
```

永久生效后重新加载：

```bash
sysctl -p
```

如果希望分文件管理，也可以新建：

```text
/etc/sysctl.d/99-tuning.conf
```

然后执行：

```bash
sysctl --system
```

## 5. 验证命令

系统参数验证：

```bash
sysctl net.core.somaxconn
sysctl net.netfilter.nf_conntrack_max
```

宿主机文件句柄验证：

```bash
ulimit -n
```

容器文件句柄验证：

```bash
docker exec -it <container_name> sh -c 'ulimit -n'
```

Caddy 配置检查：

```bash
caddy validate --config /etc/caddy/Caddyfile
```

Caddy 重载：

```bash
caddy reload --config /etc/caddy/Caddyfile
```

## 6. 风险与回滚

### 6.1 `nf_conntrack_max` 不是越大越好

连接跟踪表会占用内存。`2C4G` 小机器不建议直接套用 `262144` 或 `524288` 这类更激进的值，先以 `131072` 为起点更稳。

### 6.2 `somaxconn` 不是首字慢的直接原因

`somaxconn` 更偏向高峰接入排队能力。它能改善突发建连抖动，但不会直接加快模型响应速度。

### 6.3 首字慢优先排查 Go 中转

如果 SSE 首字仍然偏慢，优先检查：

- Go 中转是否在每个 chunk 后及时 `Flush()`
- 上游模型接口本身是否首包慢
- 是否存在额外代理层或 CDN 干预流式输出

### 6.4 回滚方法

如果调整后出现异常，可以把 `/etc/sysctl.conf` 恢复到原值，然后重新执行：

```bash
sysctl -p
```

如果是 Docker 容器配置调整导致问题，恢复 `docker-compose.yml` 后重启容器即可。

## 7. 最终建议

对 `2C4G` 小机器来说，优先采用以下组合：

- Caddy 采用 `/ws` 单独分流
- `response_header_timeout` 设为 `90s`
- Docker 显式配置 `ulimits.nofile=262144`
- `/etc/sysctl.conf` 中设置：
  - `net.core.somaxconn = 16384`
  - `net.netfilter.nf_conntrack_max = 131072`

这套参数更适合“稳中求性能”的 AI 中转机场景。
