# Changelog

## 2026-03-09 - v0.1.95

### Added
- 新增 `deploy/docker-compose.build-local.yml`，支持使用本地源码构建镜像后通过 compose 启动。
- 新增 `deploy/docker-compose.ha.yml`，提供 `sub2api-b` 第二实例（1456）用于双实例部署。
- 新增 `deploy/restart-zero-downtime.sh`，支持 `sub2api-b -> sub2api` 滚动无中断重启。
- 新增根目录 `AGENTS.md`，记录本地构建启动、重启与巡检命令。

### Changed
- `ai.xyyamsz.cn` 反代上游切换为双后端（1455/1456）负载与故障切换。
- 统一双实例健康检查为 `wget`，避免本地镜像缺少 `curl` 导致误判 `unhealthy`。

## 2026-03-09 - v0.1.94

### Added
- 新增独立的管理员令牌管理模块，并在订阅管理下方提供独立入口与标签页。
- 支持令牌管理生成账号使用 API Key 登录，账号密码登录保留为备选方式。

### Changed
- API Key 登录态下隐藏前端 API 密钥入口、快捷操作、统计卡片，并拦截直接访问 `/keys`。
- 更新管理员账号邮箱为 `511071161@qq.com`。
- 调整 Docker 健康检查实现，改为使用 `wget`。

### Fixed
- 修复令牌管理中复制 API Key 失败的问题。
- 修复令牌创建失败时的回滚流程，避免残留订阅或用户数据。
