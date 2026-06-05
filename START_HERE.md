# 🚀 START HERE - Fittingly Go Backend with Docker

**Quick Links**: [Get Started in 30s](#30-second-quickstart) • [Documentation Index](#documentation) • [Common Tasks](#common-tasks)

---

## ⚡ 30 Second Quickstart

```bash
# That's it! One command:
docker-compose up -d

# Verify it works:
curl http://localhost:8080/api/products
```

**URLs:**
- API: http://localhost:8080
- Database UI: http://localhost:8081

---

## 📚 Documentation

### New to Docker? Start Here
👉 **[GETTING_STARTED.md](GETTING_STARTED.md)** (5 min read)
- Simple, straightforward guide
- Common commands explained
- Basic troubleshooting

### Need Full Details?
👉 **[DOCKER_SETUP.md](DOCKER_SETUP.md)** (Comprehensive)
- Complete reference (400+ lines)
- Advanced topics
- Performance optimization
- Full troubleshooting guide

### Want to Know What Was Created?
👉 **[DOCKER_COMPLETE.md](DOCKER_COMPLETE.md)**
- Summary of all files
- What each file does
- Setup overview

### Quick Reference
👉 **[DOCKER_SUMMARY.txt](DOCKER_SUMMARY.txt)**
- All commands at a glance
- Quick checklist
- File locations

---

## 📋 Common Tasks

### Start Your Backend
```bash
# Option 1: Background (recommended)
docker-compose up -d

# Option 2: Foreground (see logs)
docker-compose up

# Option 3: Using Make
make docker-up-d

# Option 4: With hot-reload (development)
docker-compose -f docker-compose.yml -f docker-compose.override.yml up
```

### Monitor Services
```bash
# Check status
docker-compose ps

# View logs
docker-compose logs -f

# View API logs
docker-compose logs -f api

# Using Make
make docker-logs
```

### Stop Services
```bash
# Keep data
docker-compose stop
docker-compose down

# Delete everything
docker-compose down -v
```

### Test the API
```bash
# List products
curl http://localhost:8080/api/products

# Register user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "Password123",
    "phone": "9876543210",
    "gender": "male",
    "dateOfBirth": "1990-01-15"
  }'
```

### Access Database
```bash
# Web UI: http://localhost:8081
# Username: admin
# Password: password

# Or command line:
docker-compose exec mongodb mongosh -u root -p rootpassword kloset_dev

# Inside MongoDB shell:
db.users.find()
db.products.find()
```

### Useful Make Commands
```bash
make help                # Show all commands
make docker-up-d         # Start
make docker-down         # Stop
make docker-logs         # View logs
make db-reset            # Reset database
make docker-clean        # Clean everything
```

---

## 🏗️ Project Structure

```
Kloset/
├── docker-compose.yml            ← Start here to run
├── Dockerfile                    ← Production build
├── Dockerfile.dev                ← Development with hot-reload
├── .env.docker                   ← Configuration template
├── Makefile                      ← Shortcuts (make help)
├── GETTING_STARTED.md            ← Quick guide (read this first!)
├── DOCKER_SETUP.md               ← Full documentation
├── DOCKER_COMPLETE.md            ← What was created
├── scripts/
│   ├── quickstart.sh             ← One-command setup
│   └── init-mongo.js             ← Database init
└── [Go backend code]
```

---

## ✅ Verify It Works

After running `docker-compose up -d`, check:

```bash
# 1. Services are running
docker-compose ps
# Look for "healthy" status

# 2. API responds
curl http://localhost:8080/api/products
# Should return product list

# 3. Database is accessible
docker-compose exec mongodb mongosh -u root -p rootpassword kloset_dev -eval "db.users.countDocuments()"
# Should return a number
```

---

## 🔧 Customization

### Change Port
Edit `docker-compose.yml`, change `8080:8080` to `9000:8080`

### Change MongoDB Credentials
Edit `.env`, update `MONGO_ROOT_USER` and `MONGO_ROOT_PASSWORD`

### Add Frontend CORS
Edit `.env`, update `CORS_ORIGIN` to your frontend URL

### Add Cloudinary
Edit `.env`, add your Cloudinary credentials

---

## 🐛 Need Help?

| Problem | Solution |
|---------|----------|
| "Port already in use" | `lsof -i :8080` then `kill -9 <PID>` |
| "Docker not running" | Open Docker Desktop or run `sudo systemctl start docker` |
| "API won't connect" | Check logs: `docker-compose logs api` |
| "MongoDB error" | Check logs: `docker-compose logs mongodb` |
| Need more help? | See [DOCKER_SETUP.md](DOCKER_SETUP.md) Troubleshooting section |

---

## 📖 Learning Path

1. **This file** → Get oriented (you are here!)
2. **[GETTING_STARTED.md](GETTING_STARTED.md)** → Learn basics
3. Run `docker-compose up -d` → Start services
4. Test `curl http://localhost:8080/api/products` → Verify working
5. Open http://localhost:8081 → Explore database
6. [DOCKER_SETUP.md](DOCKER_SETUP.md) → Deep dive when ready

---

## 🎯 What You Have

✅ 45+ API endpoints ready to use
✅ MongoDB database with sample data
✅ Database web UI (MongoDB Express)
✅ Health checks on all services
✅ Persistent data volumes
✅ Production-ready Docker setup
✅ Development with hot-reload option
✅ Comprehensive documentation

---

## 🚀 Ready?

```bash
# Start now:
docker-compose up -d

# Check status:
docker-compose ps

# Test API:
curl http://localhost:8080/api/products

# Explore database:
# Open http://localhost:8081 in your browser
```

**That's it! Your backend is running!** 🎉

---

## 📚 Full Documentation Index

| File | Purpose | Read Time |
|------|---------|-----------|
| **START_HERE.md** | This file - quick orientation | 2 min |
| **GETTING_STARTED.md** | Quick reference and common tasks | 5 min |
| **DOCKER_SETUP.md** | Comprehensive guide with all details | 15 min |
| **DOCKER_COMPLETE.md** | What was created and why | 5 min |
| **DOCKER_SUMMARY.txt** | Quick reference sheet | 3 min |
| **Makefile** | All available commands | on demand |
| **IMPLEMENTATION_COMPLETE.md** | Original project completion summary | reference |
| **README.md** | General project info | reference |

---

## 💡 Pro Tips

```bash
# Rebuild and run
docker-compose up --build -d

# Follow logs while working
docker-compose logs -f api

# Quick shell in container
docker-compose exec api sh

# Quick database shell
docker-compose exec mongodb mongosh -u root -p rootpassword kloset_dev

# Total cleanup (fresh start)
docker-compose down -v && docker-compose up -d

# Check resource usage
docker stats
```

---

## 🔐 Security Notes

**Development** (current):
- Basic credentials (easy to remember)
- Open CORS to localhost
- Basic rate limiting

**Production** (before deploying):
- Generate strong secrets: `openssl rand -base64 32`
- Update all credentials in `.env.production`
- Restrict CORS to your domain
- Use environment variable injection
- Never commit `.env` to git

---

## ❓ Quick Q&A

**Q: Where do I run the commands?**
A: In your terminal, in the project directory (`cd /path/to/Kloset`)

**Q: Do I need Docker Desktop?**
A: Yes, install from docker.com

**Q: Can I change the port?**
A: Yes, edit `docker-compose.yml`

**Q: How do I see what's happening?**
A: Run `docker-compose logs -f`

**Q: How do I stop it?**
A: Run `docker-compose down`

**Q: Will I lose data if I stop?**
A: No, data is saved in Docker volumes

**Q: Can I use this in production?**
A: Yes, it's production-ready!

**Q: What about frontend?**
A: Connect by updating `CORS_ORIGIN` in `.env`

---

## 🎉 You're All Set!

Your complete backend stack is containerized and ready to use. One command to start:

```bash
docker-compose up -d
```

That's it! 🚀

---

**Need help?** See the appropriate documentation:
- Quick start? → [GETTING_STARTED.md](GETTING_STARTED.md)
- Full guide? → [DOCKER_SETUP.md](DOCKER_SETUP.md)
- Troubleshooting? → [DOCKER_SETUP.md](DOCKER_SETUP.md) (Troubleshooting section)

**Happy coding!** 💻
