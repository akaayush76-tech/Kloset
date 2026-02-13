# Docker Setup Guide - Fittingly Go Backend API

This guide will help you get the entire Fittingly Go backend application stack up and running using Docker and Docker Compose.

## Prerequisites

- Docker (version 20.10+)
- Docker Compose (version 2.0+)
- At least 4GB of free disk space

### Installation

**macOS:**
```bash
brew install docker docker-compose
# or download Docker Desktop from https://www.docker.com/products/docker-desktop
```

**Linux:**
```bash
sudo apt-get update
sudo apt-get install docker.io docker-compose
```

**Windows:**
Download Docker Desktop from https://www.docker.com/products/docker-desktop

## Quick Start

### 1. Clone/Navigate to Project

```bash
cd /path/to/Kloset
```

### 2. Start All Services

```bash
# Start in foreground (see logs)
docker-compose up

# OR start in background
docker-compose up -d
```

The first run will:
- Build the Go application image
- Start MongoDB container
- Initialize database with sample data
- Start the API server on port 8080
- Start MongoDB Express (web UI) on port 8081

### 3. Verify Services

Check if all services are running:

```bash
docker-compose ps
```

Expected output:
```
NAME                  COMMAND                  SERVICE      STATUS
kloset-api            "./server"               api          Up (healthy)
kloset-mongodb        "mongod"                 mongodb      Up (healthy)
kloset-mongo-express  "tini -- node ..."       mongo-express Up
```

### 4. Test the API

```bash
# Test API connectivity
curl http://localhost:8080/api/products

# Register a new user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "Password123",
    "phone": "9876543210",
    "gender": "male",
    "dateOfBirth": "1990-01-15"
  }'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "Password123"
  }'
```

## Service Details

### API Server
- **URL**: http://localhost:8080
- **Container**: kloset-api
- **Port**: 8080
- **Framework**: Go with Gin
- **Status Endpoint**: GET /api/products (returns 200 if healthy)

### MongoDB
- **URL**: mongodb://root:rootpassword@localhost:27017/
- **Container**: kloset-mongodb
- **Port**: 27017
- **Database**: kloset_dev
- **Credentials**: 
  - Username: `root`
  - Password: `rootpassword`
- **Volumes**: MongoDB data persisted in named volume `mongodb_data`

### MongoDB Express (Web UI)
- **URL**: http://localhost:8081
- **Container**: kloset-mongo-express
- **Port**: 8081
- **Credentials**:
  - Username: `admin`
  - Password: `password`

## Common Commands

### View Logs

```bash
# View logs from all services
docker-compose logs -f

# View logs from specific service
docker-compose logs -f api
docker-compose logs -f mongodb

# View last 100 lines
docker-compose logs --tail=100 api
```

### Stop Services

```bash
# Stop all services (keeps volumes)
docker-compose stop

# Stop and remove containers (keeps volumes)
docker-compose down

# Stop and remove everything including volumes
docker-compose down -v
```

### Rebuild Services

```bash
# Rebuild without cache
docker-compose build --no-cache

# Rebuild and restart
docker-compose up --build
```

### Access MongoDB CLI

```bash
# Connect to MongoDB directly
docker-compose exec mongodb mongosh -u root -p rootpassword kloset_dev

# Example MongoDB commands:
# Show all collections
db.getCollectionNames()

# Count documents
db.users.countDocuments()

# Find all users
db.users.find()

# Insert sample document
db.products.insertOne({name: "Test Product", price: 100})
```

### Access API Container

```bash
# Open shell in API container
docker-compose exec api sh

# View environment variables
docker-compose exec api env

# View running processes
docker-compose exec api ps aux
```

## Development Workflow

### Option 1: Development with Hot Reload

For development with automatic code reloading:

```bash
# Use docker-compose with development profile
docker-compose -f docker-compose.yml -f docker-compose.override.yml up

# This uses Dockerfile.dev with Air for hot-reload
```

The override file is automatically loaded, which:
- Uses `Dockerfile.dev` instead of `Dockerfile`
- Enables debug logging
- Mounts source code for live editing
- Includes MongoDB Express on port 8081

### Option 2: Manual Rebuild on Changes

```bash
# After code changes
docker-compose build api

# Restart the service
docker-compose up -d api
```

### Option 3: Local Development (No Docker)

If you prefer to run locally without Docker:

```bash
# Install dependencies
go mod download

# Update MongoDB URI in .env.development to localhost
MONGODB_URI=mongodb://localhost:27017

# Start MongoDB locally (if installed)
mongod

# Run the server
go run ./cmd/server/main.go
```

## Environment Configuration

### Using Custom Environment Variables

Create a `.env` file in the project root:

```bash
# Server Configuration
PORT=8080
ENV=development
LOG_LEVEL=debug

# MongoDB
MONGO_ROOT_USER=root
MONGO_ROOT_PASSWORD=rootpassword
MONGODB_DATABASE=kloset_dev

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=7

# Cloudinary (optional)
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret

# CORS
CORS_ORIGIN=http://localhost:3000

# Rate Limiting
RATE_LIMIT_WINDOW=15m
RATE_LIMIT_MAX_REQUESTS=100

# File Upload
MAX_FILE_SIZE=10485760
```

Then run:

```bash
docker-compose up
```

### For Production Deployment

Create `.env.production`:

```bash
PORT=8080
ENV=production
LOG_LEVEL=info

# Strong JWT secret
JWT_SECRET=$(openssl rand -base64 32)

# Cloudinary (required)
CLOUDINARY_CLOUD_NAME=your_production_cloud
CLOUDINARY_API_KEY=your_production_key
CLOUDINARY_API_SECRET=your_production_secret

# Production MongoDB (external or managed)
MONGODB_URI=mongodb+srv://user:password@your-cluster.mongodb.net/

# Restricted CORS
CORS_ORIGIN=https://yourdomain.com

# Stricter rate limiting
RATE_LIMIT_MAX_REQUESTS=50
```

## Database Initialization

When the MongoDB container starts, it automatically:

1. Creates the `kloset_dev` database
2. Creates all required collections:
   - users
   - products
   - orders
   - reviews
   - wardrobe_items
3. Creates all necessary indexes
4. Inserts sample data (1 demo user + 2 sample products)

### Manual Database Operations

Reset the database:

```bash
# Connect to MongoDB
docker-compose exec mongodb mongosh -u root -p rootpassword kloset_dev

# In the MongoDB shell
db.dropDatabase()
```

## Network & Port Mapping

```
┌─────────────────────────────────────────────────────┐
│           Docker Compose Network (kloset-network)  │
├─────────────────────────────────────────────────────┤
│                                                      │
│  ┌──────────────┐      ┌──────────────────┐        │
│  │  API Server  │      │    MongoDB       │        │
│  │ :8080        │ ───→ │   :27017         │        │
│  │ kloset-api   │      │  kloset-mongodb  │        │
│  └──────────────┘      └──────────────────┘        │
│                                │                     │
│  ┌──────────────────────────────────────────┐      │
│  │      MongoDB Express (Web UI)            │      │
│  │      :8081 (mongo-express)               │      │
│  │      Connects to MongoDB :27017          │      │
│  └──────────────────────────────────────────┘      │
│                                                      │
└─────────────────────────────────────────────────────┘

External Access:
- API: http://localhost:8080
- Mongo Express: http://localhost:8081
- MongoDB: mongodb://localhost:27017 (from host)
```

## Troubleshooting

### Issue: "Address already in use"

Port is already in use on your machine:

```bash
# Change ports in docker-compose.yml
# Find the "ports" section and modify:
ports:
  - "8090:8080"  # Use 8090 instead of 8080

# Or kill the process using the port:
# On macOS/Linux
lsof -i :8080
kill -9 <PID>

# On Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

### Issue: MongoDB connection refused

```bash
# Check if MongoDB is running
docker-compose ps mongodb

# View MongoDB logs
docker-compose logs mongodb

# Restart MongoDB
docker-compose restart mongodb

# Wait for health check
docker-compose logs --follow mongodb | grep "ready to accept connections"
```

### Issue: API not starting

```bash
# View API logs
docker-compose logs api

# Check environment variables
docker-compose exec api env | grep MONGODB

# Rebuild the image
docker-compose build --no-cache api
docker-compose up -d api
```

### Issue: Cannot connect to MongoDB from API

Make sure MongoDB container name is correct in connection string:

```bash
# Correct (for Docker):
MONGODB_URI=mongodb://root:rootpassword@mongodb:27017/

# Incorrect:
MONGODB_URI=mongodb://root:rootpassword@localhost:27017/
```

### Issue: Database initialization didn't run

```bash
# Clear MongoDB volume
docker-compose down -v

# Restart
docker-compose up

# Or manually run initialization:
docker-compose exec mongodb mongosh -u root -p rootpassword < scripts/init-mongo.js
```

## Performance Optimization

### Limit Memory Usage

Edit `docker-compose.yml`:

```yaml
services:
  api:
    deploy:
      resources:
        limits:
          memory: 512M
  mongodb:
    deploy:
      resources:
        limits:
          memory: 1G
```

### Enable Resource Monitoring

```bash
# Monitor resource usage
docker stats

# Watch container metrics
docker-compose stats
```

## Advanced Topics

### Multi-Stage Build Benefits

The `Dockerfile` uses multi-stage build:
- Build stage compiles Go code
- Final stage runs lightweight alpine image
- Reduces final image size from ~800MB to ~32MB

### Health Checks

All services have health checks:

```bash
# View health status
docker-compose ps

# Check specific service
docker inspect kloset-api | grep -A 10 "Health"
```

### Volume Management

```bash
# List volumes
docker volume ls

# Inspect volume
docker volume inspect kloset_mongodb_data

# Clean up unused volumes
docker volume prune
```

## Cleanup

### Remove All Containers & Volumes

```bash
# Remove stopped containers, unused networks, dangling images
docker system prune

# Also remove volumes
docker system prune -a --volumes
```

### Full Reset

```bash
docker-compose down -v --remove-orphans
docker system prune -a --volumes
docker-compose up --build
```

## Next Steps

1. **API Testing**: Use Postman or cURL to test endpoints
2. **Frontend Integration**: Update frontend CORS_ORIGIN to match your frontend URL
3. **Cloudinary Setup**: Add real Cloudinary credentials for image uploads
4. **Production Deployment**: Use docker-compose for staging/production with proper secrets management
5. **Monitoring**: Add Docker logging drivers and monitoring solutions

## Additional Resources

- [Docker Documentation](https://docs.docker.com)
- [Docker Compose Specification](https://github.com/compose-spec/compose-spec)
- [MongoDB Docker Image](https://hub.docker.com/_/mongo)
- [Go Docker Best Practices](https://docs.docker.com/language/golang/)
- [Air - Live Reload](https://github.com/cosmtrek/air)

## Support

For issues or questions:
- Check Docker logs: `docker-compose logs -f`
- Verify .env file configuration
- Ensure Docker and Docker Compose are up to date
- Check port availability on your system

---

**Happy coding! 🚀**
