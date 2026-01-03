# LMP 后端（最小可运行框架）

## 运行

```bash
cd backend
go run ./cmd/server
```

可用端点：
- GET /health
- POST /api/v1/auth/device/code
- POST /api/v1/auth/device/verify
- GET /api/v1/files  （需要 Authorization: Bearer <token>）
- GET /api/v1/browse  （需要 Authorization: Bearer <token>）
- GET /api/v1/search?q=keyword&page=1&page_size=50  （需要 Authorization: Bearer <token>）
- GET /api/v1/meta?path=<urlencoded>  （需要 Authorization: Bearer <token>）
- GET /api/v1/download?path=<urlencoded>  （需要 Authorization: Bearer <token>，支持 Range）

前端静态界面（用于 Gate-2 联调）：
- GET /ui/  打开 Web 界面进行设备认证、浏览、搜索与下载

更详细的 API 说明与示例见：docs/API.md

## 配置

在项目根或 backend 目录下创建 `config.json`（可选）：

```json
{
  "server": { "port": 8080, "lan_only": true },
  "files": { "share_dirs": [] },
  "auth": { "enable_device_auth": true, "session_timeout": "30m" },
  "rate_limit": { "enabled": true, "requests_per_second": 10, "burst": 20 }
}
```

配置支持热重载：修改 `config.json` 将在运行时自动生效，当前已支持对 `server.lan_only`、`rate_limit.*` 和 `files.share_dirs` 的动态更新；共享目录变更会自动重建搜索索引，无需重启服务。

## 阶段与验收

本项目执行“三阶段合并”流程，仅在以下 Gate 进行验收：
- Gate-1（后端稳定化与API收尾）：完成后端收尾与测试报告提交
- Gate-2（前端MVP与端到端联调）：完成前端最小功能与端到端报告
- Gate-3（部署与质保闭环）：完成部署脚本、性能与安全报告、发布材料

我将自行在阶段内部进行分阶段测试与修复，仅在 Gate-1/2/3 停下请你检验与验收。

