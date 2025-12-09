# GitHub Actions - Docker Hub Deployment

Tự động build và push Docker images lên Docker Hub khi có push/tag mới.

## Setup

### 1. Tạo Docker Hub Access Token

1. Đăng nhập vào [Docker Hub](https://hub.docker.com)
2. Vào **Account Settings** → **Security** → **New Access Token**
3. Đặt tên token: `github-actions`
4. Copy token (chỉ hiện 1 lần!)

### 2. Thêm Secrets vào GitHub Repository

1. Vào repository → **Settings** → **Secrets and variables** → **Actions**
2. Click **New repository secret**
3. Thêm 2 secrets:

```
DOCKER_HUB_USERNAME = your-dockerhub-username
DOCKER_HUB_TOKEN = your-access-token-from-step-1
```

### 3. Push Code

```bash
git add .
git commit -m "Add Docker and GitHub Actions"
git push origin main
```

## Workflows

### 1. `docker-build.yml` - Build & Push Images

**Triggers:**
- Push to `main`, `master`, `develop`
- Push tags `v*`
- Manual dispatch
- Pull requests (build only, no push)

**Builds 3 images:**
- `your-username/grpc-hub:latest`
- `your-username/grpc-webapi:latest`
- `your-username/grpc-worker:latest`

**Tags tạo tự động:**
- `latest` - từ main branch
- `v1.0.0` - từ git tag
- `main-abc1234` - từ commit SHA
- `pr-123` - từ pull request

**Platforms:**
- `linux/amd64`
- `linux/arm64`

### 2. `docker-test.yml` - Integration Tests

**Triggers:**
- Mọi push và PR

**Tests:**
- Build docker-compose
- Start all services
- Test API endpoints
- Check logs

## Usage

### Release mới

```bash
# Tag version mới
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GitHub Actions sẽ tự động build và push:
# - your-username/grpc-hub:v1.0.0
# - your-username/grpc-hub:v1.0
# - your-username/grpc-hub:v1
# - your-username/grpc-hub:latest
```

### Pull images từ Docker Hub

```bash
# Pull latest
docker pull your-username/grpc-hub:latest
docker pull your-username/grpc-webapi:latest
docker pull your-username/grpc-worker:latest

# Pull specific version
docker pull your-username/grpc-hub:v1.0.0
```

### Sử dụng với docker-compose

Tạo `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  hub:
    image: your-username/grpc-hub:latest
    ports:
      - "50051:50051"
    environment:
      - PORT=50051
    restart: unless-stopped

  worker:
    image: your-username/grpc-worker:latest
    depends_on:
      - hub
    environment:
      - HUB_ADDRESS=hub:50051
    restart: unless-stopped

  webapi:
    image: your-username/grpc-webapi:latest
    depends_on:
      - hub
    ports:
      - "8081:8081"
    environment:
      - HUB_ADDRESS=hub:50051
    restart: unless-stopped
```

Deploy:

```bash
docker-compose -f docker-compose.prod.yml up -d
```

## Monitoring GitHub Actions

### View workflow runs

1. Vào repository → **Actions** tab
2. Click vào workflow run để xem details
3. Download logs nếu cần

### Manual trigger

1. Vào **Actions** → **Build and Push Docker Images**
2. Click **Run workflow**
3. Chọn branch và click **Run workflow**

## Badges

Thêm vào README.md:

```markdown
![Docker Build](https://github.com/your-username/your-repo/workflows/Build%20and%20Push%20Docker%20Images/badge.svg)
![Docker Test](https://github.com/your-username/your-repo/workflows/Docker%20Compose%20Build%20Test/badge.svg)
```

## Multi-architecture Support

Images được build cho cả:
- **amd64**: Intel/AMD processors (x86_64)
- **arm64**: ARM processors (Apple Silicon, Raspberry Pi 4+)

Pull đúng architecture tự động:
```bash
# Tự động chọn đúng arch
docker pull your-username/grpc-hub:latest
```

## Advanced

### Build cho arch cụ thể

```yaml
platforms: linux/amd64  # Chỉ amd64
platforms: linux/arm64  # Chỉ arm64
platforms: linux/amd64,linux/arm64,linux/arm/v7  # Multi
```

### Custom registry

Thay Docker Hub bằng GitHub Container Registry:

```yaml
- name: Log in to GitHub Container Registry
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}
```

Images sẽ là: `ghcr.io/your-username/grpc-hub:latest`

## Troubleshooting

### Build failed

1. Check logs trong Actions tab
2. Test build locally:
   ```bash
   docker build -f Dockerfile.hub -t test:latest .
   ```

### Push failed

1. Verify Docker Hub credentials
2. Check token permissions
3. Re-create access token nếu cần

### Tests failed

1. View service logs trong Actions
2. Test locally:
   ```bash
   docker-compose up
   curl http://localhost:8081/api/status
   ```

## Security

### Secrets management

- ✅ Dùng GitHub Secrets cho credentials
- ✅ Token có expiry date
- ✅ Rotate tokens định kỳ
- ❌ Không commit credentials vào code

### Image scanning

Thêm vulnerability scanning:

```yaml
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: your-username/grpc-hub:latest
    format: 'sarif'
    output: 'trivy-results.sarif'
```

## Cost Optimization

### Cache optimization

- ✅ Sử dụng GitHub Actions cache
- ✅ Multi-stage builds
- ✅ Layer caching với buildx

### Build only on changes

```yaml
paths:
  - 'cmd/**'
  - 'internal/**'
  - 'services/**'
  - 'Dockerfile*'
  - 'go.mod'
  - 'go.sum'
```

## Resources

- [Docker Hub](https://hub.docker.com)
- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Docker Build Action](https://github.com/docker/build-push-action)
- [Docker Metadata Action](https://github.com/docker/metadata-action)
