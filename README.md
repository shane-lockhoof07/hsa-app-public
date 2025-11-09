# HSA Receipt Manager

A complete full-stack application for managing Health Savings Account (HSA) receipts with AI-powered OCR, intelligent deduction calculations, and comprehensive receipt tracking.

## Overview

The HSA Receipt Manager helps you organize and track medical receipts for HSA reimbursement. Upload receipt images, automatically extract information using Claude AI, mark receipts as used/unused, and find optimal receipt combinations to match your reimbursement amounts.

## Features

- 📤 **Smart Receipt Upload** - Drag-and-drop with support for JPG, PNG, HEIC, and PDF
- 🤖 **AI-Powered OCR** - Claude 3.5 Haiku extracts vendor, amount, date, and line items
- 🔍 **Duplicate Detection** - Prevents duplicate receipts via image hash and data matching
- 📊 **Receipt Management** - Edit, delete, and categorize receipts as Yes/No/Partially HSA-qualified
- 🧮 **Deduction Calculator** - Finds optimal receipt combinations using subset-sum algorithm
- 📁 **Smart Organization** - Automatically organizes receipts by year and used/unused status
- 💰 **Real-time Tracking** - See available vs. used HSA amounts at a glance

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        User's Browser                            │
│                   (Vue.js SPA + Vuetify)                        │
└────────────────────────┬────────────────────────────────────────┘
                         │ HTTP/REST API
                         │
┌────────────────────────▼────────────────────────────────────────┐
│                     Go Backend API                               │
│    • REST endpoints • File management • Business logic          │
└─────┬──────────────────┬────────────────────┬───────────────────┘
      │                  │                    │
      │                  │                    │
      ▼                  ▼                    ▼
┌──────────┐    ┌─────────────────┐    ┌──────────────┐
│PostgreSQL│    │  OCR Service    │    │ File Storage │
│ Database │    │ (Claude Vision) │    │  /data/hsa   │
└──────────┘    └─────────────────┘    └──────────────┘
```

## Technology Stack

### Frontend
- **Framework**: Vue 3 (Composition API)
- **UI Library**: Vuetify 3
- **Routing**: Vue Router 4
- **State**: Pinia
- **HTTP**: Axios
- **Build**: Vite

### Backend API
- **Language**: Go 1.21
- **Database**: PostgreSQL 14
- **Driver**: lib/pq
- **Migrations**: SQL files

### OCR Service
- **Framework**: FastAPI (Python 3.10)
- **AI Engine**: Claude 3.5 Haiku (Anthropic)
- **Image Processing**: Pillow, pillow-heif
- **PDF Support**: pdf2image

### Infrastructure
- **Containerization**: Docker
- **Orchestration**: Kubernetes (MicroK8s/K3s)
- **Web Server**: Nginx (production)
- **Database**: PostgreSQL 14

## Project Structure

```
hsa-receipt-manager/
├── api-go/                      # Go backend API
│   ├── config/                  # Configuration management
│   ├── internal/                # Core business logic
│   ├── migrations/              # Database migrations
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   └── README.md
├── frontend/                    # Vue.js frontend
│   ├── src/
│   │   ├── router/              # Vue Router config
│   │   ├── services/            # API client
│   │   ├── stores/              # Pinia stores
│   │   ├── views/               # Page components
│   │   ├── App.vue
│   │   └── main.js
│   ├── Dockerfile
│   ├── Dockerfile.dev
│   ├── nginx.conf
│   ├── package.json
│   └── README.md
├── service-ocr/                 # OCR microservice
│   ├── main.py                  # FastAPI app
│   ├── requirements.txt
│   ├── Dockerfile
│   └── README.md
├── k8s/                         # Kubernetes manifests
│   ├── namespace.yaml
│   ├── pvc.yaml
│   ├── api-deployment.yaml
│   ├── frontend-deployment.yaml
│   ├── ocr-deployment.yaml
│   ├── hsa-app-secret.yaml.example
│   ├── hsa-app-network-policy.yaml
│   └── README.md
├── docker-compose.yaml.example  # Docker Compose template
├── docker-compose.dev.yaml      # Development overrides
├── .env.example                 # Environment template
├── .gitignore
└── README.md                    # This file
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- Node.js 18+ (for local development)
- Python 3.10+ (for local development)
- PostgreSQL 14+ (or use Docker)
- Anthropic API key (for Claude)

### 1. Clone and Configure

```bash
# Clone repository
git clone <your-repo-url>
cd hsa-receipt-manager

# Create environment file
cp .env.example .env

# Edit .env and add your Claude API key
nano .env
```

### 2. Choose Your Deployment Method

#### Option A: Docker Compose (Recommended for Development)

```bash
# Copy docker-compose template
cp docker-compose.yaml.example docker-compose.yaml

# Edit docker-compose.yaml:
# - Update POSTGRES_PASSWORD
# - Ensure CLAUDE_API_KEY is set (from .env)

# Start all services
docker compose up -d

# View logs
docker compose logs -f

# Access application
# Frontend: http://localhost:3000
# API: http://localhost:8080
# OCR: http://localhost:8001
```

#### Option B: Kubernetes (Recommended for Production)

```bash
# See k8s/README.md for detailed instructions

# Quick start:
cd k8s
kubectl apply -f namespace.yaml
kubectl apply -f pvc.yaml
# ... follow k8s/README.md for complete steps

# Access application
# Frontend: http://<node-ip>:30591
# API: http://<node-ip>:30081
```

#### Option C: Local Development

```bash
# 1. Start PostgreSQL (Docker or local)
docker run -d --name postgres \
  -e POSTGRES_USER=hsa \
  -e POSTGRES_PASSWORD=postgres-password \
  -e POSTGRES_DB=hsa_app \
  -p 5432:5432 \
  postgres:14-alpine

# 2. Start OCR Service
cd service-ocr
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
export CLAUDE_API_KEY=your-key-here
uvicorn main:app --reload --port 8000

# 3. Start API (new terminal)
cd api-go
export DATABASE_URL=postgres://hsa:postgres-password@localhost:5432/hsa_app?sslmode=disable
export OCR_SERVICE_URL=http://localhost:8000
go run main.go

# 4. Start Frontend (new terminal)
cd frontend
npm install
npm run dev

# Access at http://localhost:5173
```

## Configuration

### Environment Variables

Create a `.env` file in the root directory:

```bash
# Anthropic Claude API
CLAUDE_API_KEY=your-anthropic-api-key-here
CLAUDE_MODEL=claude-3-5-haiku-20241022

# Database (for docker-compose)
POSTGRES_PASSWORD=your-secure-password-here
```

### Component-Specific Configuration

Each component has its own configuration options. See individual README files:

- **API**: [api-go/README.md](./api-go/README.md)
- **Frontend**: [frontend/README.md](./frontend/README.md)
- **OCR Service**: [service-ocr/README.md](./service-ocr/README.md)
- **Kubernetes**: [k8s/README.md](./k8s/README.md)

## Usage

### Upload a Receipt

1. Navigate to the **Upload** page
2. Drag and drop a receipt image or click to browse
3. Supported formats: JPG, PNG, HEIC, PDF
4. Wait for AI processing (~2-4 seconds)
5. Review and edit extracted information if needed

### Manage Receipts

1. Go to **My Receipts** page
2. View all receipts in a sortable table
3. Click any receipt to:
   - View the original image
   - Edit vendor, amount, date
   - Change HSA qualification status (Yes/No/Partially)
   - Mark as used with a reason
   - Delete the receipt

### Calculate HSA Deduction

1. Visit the **HSA Deduction** page
2. Enter target reimbursement amount
3. Click **Calculate Optimal Receipts**
4. Review the suggested combination
5. Click **Approve & Mark as Used** to finalize

### HSA Status Options

- **Yes**: Fully HSA-qualified expenses
- **No**: Not HSA-qualified (regular purchases)
- **Partially**: Mixed receipt (calculate qualified portion manually)

## Receipt Organization

Receipts are automatically organized in the file system:

```
/data/hsa/
├── unused/
│   ├── 2024/
│   │   ├── 1699564322_walgreens_receipt.jpg
│   │   └── 1699565789_cvs_receipt.pdf
│   └── 2025/
│       └── 1704821123_target_receipt.jpg
└── used/
    ├── 2024/
    │   └── 1699234567_costco_receipt.jpg
    └── 2025/
        └── 1704456789_walmart_receipt.png
```

When a receipt is marked as "used", it's automatically moved to the appropriate `used/YEAR/` directory.

## API Endpoints

### Receipt Management
- `POST /api/receipts/upload` - Upload new receipt
- `GET /api/receipts` - List all receipts
- `GET /api/receipts/{id}` - Get specific receipt
- `PUT /api/receipts/{id}` - Update receipt
- `DELETE /api/receipts/{id}` - Delete receipt

### Deduction Calculator
- `POST /api/receipts/deduct` - Calculate optimal receipt combination

### File Serving
- `GET /receipts/file/{id}` - Serve receipt image

### Health Check
- `GET /api/health` - API health status

See [api-go/README.md](./api-go/README.md) for detailed API documentation.

## Development

### Hot Reload Development

Use Docker Compose with development overrides:

```bash
docker compose -f docker-compose.yaml -f docker-compose.dev.yaml up
```

This provides:
- Frontend hot reload (Vite HMR)
- OCR service auto-reload
- Source code mounted as volumes

### Running Tests

```bash
# Go API tests
cd api-go
go test ./...

# Frontend tests (when implemented)
cd frontend
npm run test

# OCR service tests (when implemented)
cd service-ocr
pytest
```

### Building Images

```bash
# Build all images
docker compose build

# Build specific service
docker compose build api-go

# Build for different architecture
docker buildx build --platform linux/amd64,linux/arm64 -t your-registry/hsa-app-api:latest ./api-go
```

## Deployment

### Docker Compose Production

```bash
# Use production profile
docker compose --profile prod up -d

# Access at http://localhost:3000
```

### Kubernetes Production

See [k8s/README.md](./k8s/README.md) for complete deployment guide.

Quick summary:
1. Create namespace and secrets
2. Deploy database (or use existing)
3. Deploy OCR service
4. Deploy API
5. Deploy frontend
6. Configure Ingress (optional)

### Updating Services

```bash
# Docker Compose
docker compose pull
docker compose up -d

# Kubernetes
kubectl rollout restart deployment/api-go -n hsa-app
kubectl rollout restart deployment/ocr-service -n hsa-app
kubectl rollout restart deployment/frontend -n hsa-app
```

## Database Schema

The application uses a single `receipts` table:

```sql
CREATE TABLE receipts (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    vendor VARCHAR(255),
    total_amount DECIMAL(10,2) NOT NULL,
    date DATE,
    hsa_qualified BOOLEAN DEFAULT true,
    hsa_status VARCHAR(20) DEFAULT 'Yes',
    image_path TEXT,
    image_hash VARCHAR(64),
    raw_text TEXT,
    used BOOLEAN DEFAULT false,
    used_date TIMESTAMP,
    use_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Migrations run automatically on API startup.

## Cost Analysis

### Claude API Usage

Using Claude 3.5 Haiku:
- **Cost per receipt**: ~$0.002-0.003
- **100 receipts/month**: ~$0.20-0.30
- **1,000 receipts/month**: ~$2-3
- **Alternative**: Claude 3.5 Sonnet (~$0.015/receipt, higher accuracy)

### Infrastructure

- **Docker Compose**: No additional costs (local hosting)
- **Kubernetes**: Minimal (runs on Raspberry Pi or similar)
- **PostgreSQL**: No cost (self-hosted)
- **Storage**: ~100KB per receipt average

## Security Best Practices

### Development
- Never commit `.env` files
- Use `.env.example` for templates
- Rotate API keys regularly
- Use strong database passwords

### Production
- Enable HTTPS with TLS certificates
- Use Kubernetes secrets for sensitive data
- Implement authentication (currently single-user)
- Regular security updates
- Enable database SSL connections
- Set up network policies
- Use private container registries

## Troubleshooting

### Common Issues

**Port conflicts**:
```bash
# Check what's using a port
lsof -i :8080

# Change port in docker-compose.yaml or .env
```

**OCR service fails**:
```bash
# Check Claude API key is set
docker compose logs ocr-service

# Verify key: echo $CLAUDE_API_KEY
```

**Database connection refused**:
```bash
# Check postgres is running
docker compose ps postgres

# View postgres logs
docker compose logs postgres
```

**Images not displaying**:
```bash
# Check file permissions
ls -la /data/hsa

# Verify mount in container
docker compose exec api-go ls -la /data/hsa
```

### Debug Mode

Enable detailed logging:

```bash
# API
export LOG_LEVEL=debug

# Frontend
export VITE_DEBUG=true

# OCR Service
export LOG_LEVEL=DEBUG
```

## Performance Optimization

### Frontend
- Lazy loading for routes
- Image optimization
- Caching strategy via nginx

### API
- Database connection pooling
- Efficient SQL queries with indexes
- Concurrent file operations

### OCR Service
- Rate limiting for Claude API
- Retry logic with exponential backoff
- Image preprocessing for better accuracy

## Monitoring

### Health Checks

```bash
# Check all services
curl http://localhost:8080/api/health
curl http://localhost:8001/health

# Kubernetes
kubectl get pods -n hsa-app
```

### Logs

```bash
# Docker Compose
docker compose logs -f api-go
docker compose logs -f ocr-service

# Kubernetes
kubectl logs -f deployment/api-go -n hsa-app
kubectl logs -f deployment/ocr-service -n hsa-app
```

## Backup and Recovery

### Database Backup

```bash
# Docker Compose
docker compose exec postgres pg_dump -U hsa hsa_app > backup.sql

# Restore
docker compose exec -T postgres psql -U hsa hsa_app < backup.sql
```

### File Backup

```bash
# Backup receipt files
tar -czf receipts_backup.tar.gz /data/hsa

# Restore
tar -xzf receipts_backup.tar.gz -C /
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Roadmap

### Planned Features
- [ ] Multi-user support with authentication
- [ ] Receipt categories and tags
- [ ] Export to CSV/PDF reports
- [ ] Mobile app (React Native)
- [ ] Receipt search and filtering
- [ ] Spending analytics and trends
- [ ] Email receipt forwarding
- [ ] Recurring expense tracking
- [ ] Tax year summaries
- [ ] Integration with HSA providers
- [ ] Scheduled backups to Google Drive

### Under Consideration
- [ ] Receipt splitting for shared expenses
- [ ] OCR confidence scores
- [ ] Automatic vendor categorization
- [ ] Budget tracking
- [ ] Receipt expiration warnings

## Support

For issues, questions, or contributions:

- **Documentation**: See component-specific READMEs
- **Issues**: [GitHub Issues](your-repo-url/issues)
- **Discussions**: [GitHub Discussions](your-repo-url/discussions)

## Acknowledgments

- **Anthropic** - Claude AI vision capabilities
- **Vuetify** - Material Design components
- **FastAPI** - Modern Python web framework
- **PostgreSQL** - Robust database system

## Project Status

**Current Version**: 1.0.0  
**Status**: Active Development  
**Last Updated**: November 2025

---

Built with ❤️ for better HSA management