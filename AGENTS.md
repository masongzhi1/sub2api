# 部署命令

## 启动（必须先构建本地镜像）
```bash
cd /home/ma/sub2api/deploy
cp -n .env.example .env

sudo docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.ha.yml \
  -f docker-compose.build-local.yml \
  --env-file .env \
  build sub2api sub2api-b

sudo docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.ha.yml \
  -f docker-compose.build-local.yml \
  --env-file .env \
  up -d
```

## 重启（必须先构建本地镜像，再无中断滚动重启）
```bash
cd /home/ma/sub2api/deploy

sudo docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.ha.yml \
  -f docker-compose.build-local.yml \
  --env-file .env \
  build sub2api sub2api-b

sudo /home/ma/sub2api/deploy/restart-zero-downtime.sh
```

## 常用检查命令
```bash
sudo docker compose -f /home/ma/sub2api/deploy/docker-compose.local.yml -f /home/ma/sub2api/deploy/docker-compose.ha.yml -f /home/ma/sub2api/deploy/docker-compose.build-local.yml --env-file /home/ma/sub2api/deploy/.env ps
curl -sS https://ai.xyyamsz.cn/health
```
