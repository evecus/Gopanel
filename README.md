# GoPanel 🖥️

> 轻量级服务器监控面板 · Go + Vue 3 · 支持 Linux amd64 / arm64

## ✨ 功能

- **实时监控** - CPU、内存、磁盘、网络，WebSocket 实时推送，5s 刷新
- **系统信息** - 主机名、OS、内核版本、架构、CPU 型号、运行时间
- **历史趋势** - SQLite 存储 7 天数据，图表展示
- **进程管理** - 进程列表、排序、Kill 进程
- **Docker** - 容器列表、状态、CPU/内存、启停、日志
- **Systemd** - 服务列表、启停重启、journalctl 日志
- **认证** - 密码登录

## 🚀 快速开始

```bash
# 从 Release 下载
wget https://github.com/evecus/Gopanel/releases/latest/download/gopanel-linux-amd64.tar.gz
tar -xzf gopanel-linux-amd64.tar.gz

# 运行（默认端口 1080）
./gopanel-linux-amd64 -config config.yaml
```

访问 `http://IP:1080`，默认账号：`admin` / `admin`

## ⚙️ 配置

```yaml
listen: "0.0.0.0:1080"
db_path: "gopanel.db"
collect_interval: "5s"
jwt_secret: "change-this-to-random-string"
username: "admin"
password: "admin"
alert:
  cpu: 90
  memory: 90
  disk: 90
  webhook: ""
```

## 🔨 自行构建

需要：Go 1.21+、Node.js 20+

```bash
git clone https://github.com/evecus/Gopanel
cd gopanel
make build          # 当前平台
make build-all      # amd64 + arm64
```

## 📦 作为系统服务

```bash
sudo cp gopanel-linux-amd64 /opt/gopanel/gopanel
sudo cp config.yaml /opt/gopanel/
sudo cp gopanel.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now gopanel
```

## 🔒 安全建议

- 修改默认密码
- 通过 Nginx 反代并开启 HTTPS
- 建议仅局域网访问或加 VPN

## 资源占用

- 内存：~15-30 MB
- CPU：< 0.5%（5s 采集间隔）
- 磁盘：< 50 MB（7天历史数据）

## License

MIT
