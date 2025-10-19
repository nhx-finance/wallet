# NHX Wallet Deployment Guide

## 1. Build and start containers

```bash
docker-compose -f docker-compose.prod.yml up -d --build
```

## 2. View Logs

```bash
sudo docker compose -f docker-compose.prod.yml logs --tail=10 backend
```

## 3. Pull & Restart

```bash
git pull
sudo docker compose -f docker-compose.prod.yml down
sudo docker compose -f docker-compose.prod.yml up -d
```

## 3. Stop Containers

```bash
docker-compose -f docker-compose.prod.yml down
```
