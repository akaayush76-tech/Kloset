# Getting Started - Fittingly Go Backend with Docker

**Status**: ✅ Production Ready | **Setup Time**: ~5 minutes | **Prerequisites**: Docker only

---

## 🚀 Start in 30 Seconds

```bash
# Navigate to project
cd /path/to/Kloset

# One command to run everything
docker-compose up -d

# Verify it's working
curl http://localhost:8080/api/products
```

**That's it!** Your entire backend stack is now running.

---

## 📍 Service URLs

| Service | URL | Purpose |
|---------|-----|---------|
| **API** | http://localhost:8080 | Your Go backend API |
| **MongoDB Express** | http://localhost:8081 | Database web UI |
| **MongoDB** | mongodb://root:rootpassword@localhost:27017 | Direct connection |

---

## ✅ What You Get

```
✓ Go Backend API (45+ endpoints)
✓ MongoDB Database (auto-initialized)
✓ MongoDB Web UI (for inspection)
✓ Sample data (demo user + 2 products)
✓ Health checks (all services monitored)
✓ Persistent volumes (data survives restarts)
✓ Isolated network (secure inter-service communication)
```

---

## 📚 Common Tasks

### Start Services

```bash
# Start in foreground (see logs)
docker-compose up

# Start in background
docker-compose up -d

# Start with hot-reload (development)
docker-compose -f docker-compose.yml -f docker-compose.override.yml up
```

### Monitor Services

```bash
# Check if services are running
docker-compose ps

# View all logs
docker-compose logs -f

# View API logs only
docker-compose logs -f api

# View MongoDB logs only
docker-compose logs -f mongodb
```

### Stop Services

```bash
# Stop (keep containers)
docker-compose stop

# Stop and remove containers (keep volumes)
docker-compose down

# Stop and remove everything
docker-compose down -v
```

### Test the API

```bash
# List products
curl http://localhost:8080/api/products

# Register new user
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

### Access Database

```bash
# Open MongoDB shell
docker-compose exec mongodb mongosh -u root -p rootpassword kloset_dev

# Inside MongoDB shell, try:
db.users.find()
db.products.find()
db.getCollectionNames()
```

### View Database in Web UI

Open browser: **http://localhost:8081**
- Username: `admin`
- Password: `password`

---

## 🛠️ Using Make Commands

If you prefer shortcuts, use `make`:

```bash
make help                  # Show all commands
make docker-up-d           # Start in background
make docker-down           # Stop services
make docker-logs           # View logs
make docker-logs-api       # View API logs
make docker-ps             # Show running containers
make docker-clean          # Remove everything
make db-reset              # Reset database
```

---

## 🔧 Customization

### Change API Port

Edit `docker-compose.yml`:
```yaml
api:
  ports:
    - "9000:8080"  # Now access at localhost:9000
```

### Change MongoDB Credentials

Edit `.env`:
```env
MONGO_ROOT_USER=admin
MONGO_ROOT_PASSWORD=mysecurepassword
```

### Add Frontend CORS

Edit `.env`:
```env
CORS_ORIGIN=http://localhost:3000,http://localhost:3001,http://localhost:5173
```

### Configure Cloudinary (Optional)

Edit `.env`:
```env
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
```

---

## 📖 Documentation

| Document | Purpose |
|----------|---------|
| **DOCKER_SETUP.md** | Comprehensive 400+ line guide with troubleshooting |
| **DOCKER_COMPLETE.md** | Summary of what was created |
| **Makefile** | All available commands (`make help`) |
| **GETTING_STARTED.md** | This file - quick reference |

---

## 🐛 Troubleshooting

### "Address already in use"
```bash
# Find what's using port 8080
lsof -i :8080

# Kill the process (replace <PID>)
kill -9 <PID>

# Or change port in docker-compose.yml to 9000:8080
```

### "Docker daemon not running"
- **macOS/Windows**: Open Docker Desktop
- **Linux**: `sudo systemctl start docker`

### "API won't connect"
```bash
# Check if container is running
docker-compose ps

# View error logs
docker-compose logs api

# Rebuild and restart
docker-compose build --no-cache api
docker-compose up -d api
```

### "MongoDB connection error"
```bash
# Check MongoDB is healthy
docker-compose logs mongodb

# Wait for health check to pass
sleep 10
docker-compose ps mongodb

# If still failing, reinitialize
docker-compose down -v
docker-compose up -d
```

---

## 📊 Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Docker Host                         │
├─────────────────────────────────────────────────────┤
│                                                      │
│  ┌──────────────────────────────────────────┐      │
│  │        kloset-network (Docker)           │      │
│  │                                           │      │
│  │  ┌────────────────┐  ┌────────────────┐  │      │
│  │  │   API Server   │  │    MongoDB     │  │      │
│  │  │  (Go + Gin)    │→ │  (Port 27017)  │  │      │
│  │  │  (Port 8080)   │  │ kloset_mongodb │  │      │
│  │  └────────────────┘  └────────────────┘  │      │
│  │         ↑                                  │      │
│  │         │                                  │      │
│  │  ┌────────────────┐                       │      │
│  │  │ Mongo Express  │                       │      │
│  │  │  (Port 8081)   │→ MongoDB              │      │
│  │  │ Web UI Console │                       │      │
│  │  └────────────────┘                       │      │
│  │                                            │      │
│  └────────────────────────────────────────────┘      │
│                                                      │
│  Volumes:                                           │
│  • mongodb_data   (persists DB)                     │
│  • mongodb_config (persists config)                 │
│                                                      │
└─────────────────────────────────────────────────────┘
       ↑
       │ Network Bridge
       ↓
┌─────────────────────┐
│   Your Computer     │
├─────────────────────┤
│ localhost:8080  API │
│ localhost:8081  UI  │
│ localhost:27017 DB  │
└─────────────────────┘
```

---

## 🔐 Security Notes

### Current Setup (Development)
✓ Good for: Local development and testing
✓ Credentials: Basic, easy to remember

### For Production
⚠️ Generate strong secrets:
```bash
# Generate JWT secret
openssl rand -base64 32

# Generate MongoDB password
openssl rand -base64 16

# Update .env.production with real credentials
# Never commit .env to git
```

---

## 🎯 Next Steps

### 1. Develop Locally
```bash
docker-compose up -d
# Your API is running at localhost:8080
# Make code changes and rebuild as needed
```

### 2. Test Endpoints
```bash
curl http://localhost:8080/api/products
curl -X POST http://localhost:8080/api/auth/register ...
```

### 3. Connect Frontend
```bash
# Update CORS in .env
CORS_ORIGIN=http://localhost:3000

# Restart API
docker-compose restart api
```

### 4. Deploy to Cloud
```bash
# Docker images work on any cloud platform:
# - AWS ECR
# - Google Cloud Run
# - Digital Ocean App Platform
# - Heroku
# - etc.
```

---

## 📞 File Structure

```
Kloset/
├── docker-compose.yml         ← Main config
├── .env.development           ← Dev variables
├── Dockerfile                 ← Production build
├── Dockerfile.dev             ← Dev build (hot-reload)
├── Makefile                   ← Command shortcuts
├── scripts/
│   ├── quickstart.sh          ← One-command setup
│   └── init-mongo.js          ← DB initialization
├── DOCKER_SETUP.md            ← Full guide
├── DOCKER_COMPLETE.md         ← What was created
├── GETTING_STARTED.md         ← This file
└── [Go backend code]
```

---

## ⚡ Performance Tips

### Faster Builds
```bash
# Use --no-cache only when needed
docker-compose build api

# Not
docker-compose build --no-cache api
```

### Faster Startup
```bash
# Start in background immediately
docker-compose up -d

# Don't wait for logs
docker-compose up -d api
```

### Monitor Resources
```bash
# See CPU, memory, network usage
docker stats

# Or with compose
docker-compose stats
```

---

## 🎓 Learning Path

1. **Start here**: Run `docker-compose up -d`
2. **Test API**: `curl http://localhost:8080/api/products`
3. **Explore Database**: Open http://localhost:8081
4. **Read logs**: `docker-compose logs -f`
5. **Stop & inspect**: `docker-compose down`
6. **Advanced**: See DOCKER_SETUP.md

---

## 💡 Pro Tips

```bash
# Rebuild and run in one command
docker-compose up --build -d

# Follow logs while developing
docker-compose logs -f api

# Quick shell access
docker-compose exec api sh

# Quick database access
docker-compose exec mongodb mongosh -u root -p rootpassword kloset_dev

# Total cleanup (fresh start)
docker-compose down -v && docker-compose up -d

# View environment variables
docker-compose exec api env | grep -E "(MONGODB|JWT|CORS)"
```

---

## ✨ Features Ready to Use

- ✅ User registration & login
- ✅ Product catalog with search
- ✅ Shopping cart
- ✅ Order management
- ✅ Reviews & ratings
- ✅ Virtual wardrobe
- ✅ Image uploads
- ✅ JWT authentication
- ✅ Rate limiting
- ✅ CORS handling

---

## 🆘 Quick Help

```bash
# Something not working?
docker-compose ps              # Check status
docker-compose logs api        # View API errors
docker-compose logs mongodb    # View DB errors

# Start fresh
docker-compose down -v
docker-compose up -d

# Need more help?
See DOCKER_SETUP.md for full troubleshooting guide
```

---

## 🎉 You're Ready!

```bash
# Run this one command:
docker-compose up -d

# Then visit:
# API: http://localhost:8080/api/products
# UI:  http://localhost:8081

# That's it! 🚀
```

---

**Questions?** See DOCKER_SETUP.md for detailed documentation.

**Happy coding!** 💻

