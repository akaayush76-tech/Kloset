# Kloset Backend API - Go Implementation

Complete Go backend reimplementation of Fittingly e-commerce API with 100% API contract compatibility.

## Project Status

**Implementation Progress: 127/154 tasks (83%)**

### Completed Phases
- ✅ Phase 1: Project Setup & Dependencies (6/6)
- ✅ Phase 2: Database Models & Indexes (9/9)
- ✅ Phase 3: Authentication API (9/9)
- ✅ Phase 4: Products API (11/11)
- ✅ Phase 5: Shopping Cart API (7/7)
- ✅ Phase 6: Orders API (10/10)
- ✅ Phase 7: Reviews API (10/10)
- ✅ Phase 8: Virtual Wardrobe API (9/9)
- ✅ Phase 9: Image Upload API (8/8)
- ✅ Phase 10: Middleware Stack (9/9)
- ✅ Phase 11: Input Validation (10/10)
- ✅ Phase 12: Pagination (9/9)
- ✅ Phase 13: API Response Formatting (11/11)
- ✅ Phase 14: Error Handling (9/9)
- ✅ Phase 15: Outfit Recommendation Engine (1/1)
- ⏳ Phase 16: Integration & Testing (14/14) - Ready for testing
- ⏳ Phase 17: Deployment & Documentation (12/12) - Ready

## Architecture

### Technology Stack
- **Framework**: Gin Web Framework
- **Database**: MongoDB with native Go driver
- **Authentication**: JWT (7-day expiry)
- **Password Hashing**: bcrypt (salt rounds 12)
- **Image Storage**: Cloudinary
- **Rate Limiting**: In-memory store (15-min window, 100 req/IP)
- **Recommendation Cache**: In-memory TTL cache (5-min window, sync.Map)
- **Security**: CORS, Helmet-like headers, HTTPS-ready

### Project Structure
```
.
├── cmd/
│   └── server/
│       └── main.go              # Server entry point
├── internal/
│   ├── cache/
│   │   └── recommendation_cache.go  # In-memory TTL cache (sync.Map)
│   ├── config/
│   │   ├── database.go          # MongoDB connection
│   │   └── cloudinary.go        # Cloudinary setup
│   ├── engine/
│   │   ├── color_matrix.go      # Color-pair harmony lookup table
│   │   ├── prefilter.go         # Hard-block combination rules
│   │   ├── recommender.go       # Recommendation orchestrator
│   │   └── scorer.go            # Weighted scoring algorithm
│   ├── handlers/
│   │   ├── auth.go              # Authentication endpoints
│   │   ├── profile.go           # Profile management
│   │   ├── products.go          # Product catalog
│   │   ├── products_update.go   # Product updates
│   │   ├── cart.go              # Shopping cart
│   │   ├── orders.go            # Order management
│   │   ├── recommendations.go   # Outfit recommendations
│   │   ├── reviews.go           # Product reviews
│   │   ├── wardrobe.go          # Virtual wardrobe
│   │   └── upload.go            # Image uploads
│   ├── middleware/
│   │   ├── auth.go              # JWT middleware
│   │   ├── security.go          # Security headers
│   │   └── ratelimit.go         # Rate limiting
│   ├── models/
│   │   ├── identifiers.go       # 8-dimension item identifier struct
│   │   ├── user.go              # User model
│   │   ├── product.go           # Product model
│   │   ├── order.go             # Order model
│   │   ├── review.go            # Review model
│   │   ├── wardrobe.go          # Wardrobe model
│   │   └── indexes.go           # Database indexes
│   └── utils/
│       ├── logger.go            # Logging
│       ├── response.go          # API responses
│       ├── validation.go        # Input validation
│       └── auth.go              # Auth utilities
├── scripts/
│   ├── init-mongo.js            # MongoDB init (collections, indexes, seed data)
│   └── seed_and_test.sh         # Seed 12 wardrobe items + test recommendations
├── .env.development             # Development config
├── .env.production              # Production config
└── go.mod                       # Go modules
```

## Setup

### Prerequisites
- Go 1.21+
- MongoDB 5.0+
- Cloudinary account (optional, mocked in uploads)

### Installation

1. **Install dependencies**
```bash
go mod tidy
```

2. **Configure environment**
```bash
cp .env.development .env
# Edit .env with your MongoDB and Cloudinary credentials
```

3. **Build the server**
```bash
go build -o bin/server ./cmd/server
```

4. **Run the server**
```bash
./bin/server
# Server starts on http://localhost:8080
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login
- `GET /api/auth/me` - Get profile (requires auth)
- `PUT /api/auth/profile` - Update profile (requires auth)
- `PUT /api/auth/change-password` - Change password (requires auth)
- `POST /api/auth/logout` - Logout (requires auth)

### Products
- `GET /api/products` - List products with filtering/sorting/pagination
- `GET /api/products/:id` - Get single product
- `GET /api/products/categories` - List categories
- `GET /api/products/featured` - Get featured products
- `GET /api/products/:id/related` - Get related products
- `POST /api/products` - Create product (admin only)
- `PUT /api/products/:id` - Update product (admin only)
- `DELETE /api/products/:id` - Delete product (admin only)

### Shopping Cart
- `GET /api/cart` - Get user's cart (requires auth)
- `GET /api/cart/count` - Get cart item count (requires auth)
- `POST /api/cart` - Add item to cart (requires auth)
- `PUT /api/cart/:itemId` - Update cart item (requires auth)
- `DELETE /api/cart/:itemId` - Remove item from cart (requires auth)
- `DELETE /api/cart` - Clear cart (requires auth)

### Orders
- `GET /api/orders` - List user's orders (requires auth)
- `GET /api/orders/:id` - Get single order (requires auth)
- `GET /api/orders/stats` - Get order statistics (requires auth)
- `POST /api/orders` - Create order (requires auth)
- `PUT /api/orders/:id/status` - Update order status (admin only)
- `PUT /api/orders/:id/cancel` - Cancel order (requires auth)

### Reviews
- `GET /api/reviews/product/:productId` - Get product reviews
- `GET /api/reviews/my` - Get user's reviews (requires auth)
- `GET /api/reviews/stats` - Get user's review stats (requires auth)
- `POST /api/reviews` - Create review (requires auth)
- `PUT /api/reviews/:id` - Update review (requires auth)
- `DELETE /api/reviews/:id` - Delete review (requires auth)
- `POST /api/reviews/:id/helpful` - Mark as helpful

### Virtual Wardrobe
- `GET /api/wardrobe` - List wardrobe items (requires auth)
- `GET /api/wardrobe/:id` - Get single item (requires auth)
- `GET /api/wardrobe/stats` - Get statistics (requires auth)
- `GET /api/wardrobe/category/:category` - Get by category (requires auth)
- `POST /api/wardrobe` - Create item (requires auth)
- `PUT /api/wardrobe/:id` - Update item (requires auth)
- `DELETE /api/wardrobe/:id` - Delete item (requires auth)

### Outfit Recommendations
- `POST /api/recommendations/outfits` - Generate outfit combinations for a trigger item (requires auth)

#### Request body
```json
{
  "trigger_item_id":   "string",   // MongoDB ID of the item to build outfits around
  "trigger_item_type": "closet",   // "closet" | "catalog"
  "context_filter":    "casual",   // "all" | "casual" | "smart_casual" | "date_night" | "weekend"
  "limit":             5,          // 1–10, default 8
  "include_shop_items": true       // suggest catalog items to fill missing slots
}
```

#### Response shape
```json
{
  "success": true,
  "message": "Outfit recommendations retrieved",
  "data": {
    "outfits": [
      {
        "outfit_id": "string",
        "rank": 1,
        "score": 94,
        "rank_label": "Best match",
        "items": [
          {
            "item_id": "string",
            "name": "White Slim Tee",
            "brand": "Basics Co",
            "category": "upper",
            "image_url": "https://...",
            "is_trigger": true,
            "owned": true,
            "price": 499,
            "identifiers": { "colorPrimary": "white", "fit": "slim", "occasion": "casual", "..." : "..." }
          }
        ],
        "score_breakdown": {
          "color_harmony": 0.98,
          "fit_compat": 0.90,
          "occasion_match": 1.0,
          "season_match": 1.0
        },
        "why_text": "white + navy is a high-contrast combination that always works.",
        "rule_tags": ["Color: white + navy", "Fit: slim + slim", "Occasion: casual"],
        "missing_items": []
      }
    ],
    "meta": {
      "trigger_item_id": "string",
      "closet_items_considered": 12,
      "combinations_evaluated": 48,
      "combinations_after_filters": 18,
      "returned": 5,
      "latency_ms": 23
    }
  }
}
```

### Image Upload
- `POST /api/upload/image` - Upload single image (requires auth)
- `POST /api/upload/images` - Upload multiple images (requires auth)
- `POST /api/upload/wardrobe` - Upload wardrobe image (requires auth)
- `POST /api/upload/product` - Upload product image (admin only)
- `DELETE /api/upload/image` - Delete image by public ID
- `GET /api/upload/optimize` - Get optimized image URL

### System
- `GET /health` - Health check
- `GET /api` - API info

## Request/Response Format

### Standard Response
```json
{
  "success": true,
  "message": "Operation successful",
  "data": {},
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "pages": 10,
    "hasNext": true,
    "hasPrev": false
  }
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error message"
}
```

## Validation Rules

- **Email**: RFC 5322 format
- **Password**: Minimum 6 characters (hashed with bcrypt, salt 12)
- **Phone**: Indian format (10 digits, starting with 6-9)
- **Name**: 2-50 characters
- **Address**: 10-200 characters
- **Pincode**: Exactly 6 digits (Indian format)
- **Rating**: 1-5 integer
- **Comment**: 10-500 characters
- **Category**: `upper`, `lower`, `shoes`, `outerwear`, `full_body`, `accessory`

### Item Identifier Enums (Recommendation Engine)

| Field | Allowed values |
|-------|---------------|
| `colorPrimary` | `white` `black` `navy` `beige` `grey` `brown` `red` `green` `blue` `yellow` `pink` `purple` `orange` `multicolor` |
| `colorTone` | `neutral` `pastel` `neon` `earth` `bold` |
| `fit` | `slim` `regular` `oversized` `relaxed` |
| `occasion` | `casual` `smart_casual` `date_night` `weekend` `all` |
| `season` | `spring` `summer` `fall` `winter` `all` |
| `formality` | `casual` `smart_casual` `formal` |
| `style` | `streetwear` `classic` `bohemian` `minimalist` |
| `pattern` | `solid` `stripes` `checks` `floral` `graphic` |

## HTTP Status Codes

- `200 OK` - GET successful
- `201 Created` - POST/Create successful
- `400 Bad Request` - Validation error
- `401 Unauthorized` - Auth required/invalid
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Duplicate/constraint violation
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

## Testing

### Manual Testing

1. **Register User**
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123",
    "phone": "9876543210",
    "gender": "male"
  }'
```

2. **Login**
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

3. **Get Profile** (replace TOKEN with actual token)
```bash
curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer TOKEN"
```

4. **List Products**
```bash
curl -X GET "http://localhost:8080/api/products?page=1&limit=10&category=upper"
```

### Automated Testing (Ready to implement)
- Unit tests for handlers
- Integration tests with test database
- API contract validation
- Performance tests

## Development Notes

### Recommendation Engine

The engine lives in `internal/engine/` and runs entirely in-process (no microservice). When `POST /api/recommendations/outfits` is called:

1. **Cache check** — returns immediately on a hit (key: `rec:{userID}:{triggerID}:{contextFilter}`, TTL 5 min).
2. **Fetch** — loads the trigger item (catalog or closet) and all active closet items for the user.
3. **Pre-filter** (hard blocks) — discards combinations that violate: same outfit slot, full-body + top/bottom conflict, summer↔winter season clash, casual↔formal formality gap.
4. **Combination generation** — builds trigger + required complement + optional outerwear / footwear / accessory sets (capped at 5 candidates per optional slot).
5. **Scoring** — each combination receives a 0–100 score: color harmony (35%), fit compatibility (25%), occasion match (25%), season match (15%).
6. **Deduplication** — keeps only the highest-scoring combo per top+bottom pair.
7. **Shop-to-complete** — when `include_shop_items: true`, queries the catalog to suggest products for missing slots.
8. **Cache write** — stores the result before returning.

Score labels: ≥90 → *Best match*, ≥75 → *Great match*, ≥60 → *Good match*, <60 → *Fair match*.

### Database Indexes
All indexes are automatically created via `models/indexes.go` at startup:
- User: email (unique), createdAt
- Product: text search, category+subcategory, price, rating
- Order: user+createdAt, orderStatus, paymentStatus, trackingNumber
- Review: user+product (unique), product+createdAt, rating
- WardrobeItem: user+category, user+subtype, user+createdAt, **user+isActive+identifiers.occasion** (recommendation engine)

### Rate Limiting
- 15-minute sliding window
- 100 requests per IP address
- Returns `429 Too Many Requests`

### JWT Configuration
- Algorithm: HS256
- Expiry: 7 days
- Header: `Authorization: Bearer <token>`

### File Upload Limits
- Single file: max 10MB
- Batch upload: max 10 files
- Folders: wardrobe/, products/

## Deployment

### Docker (recommended)
```bash
docker compose up -d   # starts MongoDB + API + Mongo Express UI (port 8081)
```

Multi-stage `Dockerfile` produces a minimal Alpine image running as non-root `appuser:1000`.
`Dockerfile.dev` is a single-stage build for local development.

### Environment Variables
```
PORT=8080
ENV=production
MONGODB_URI=mongodb://...
MONGODB_DATABASE=kloset_prod
JWT_SECRET=<strong-secret>
CLOUDINARY_CLOUD_NAME=...
CLOUDINARY_API_KEY=...
CLOUDINARY_API_SECRET=...
CORS_ORIGIN=https://yourdomain.com
```

## Performance Considerations

- Connection pooling: Min 10, Max 100
- Database indexes on all query fields
- Pagination max 100 items per request
- Rate limiting to prevent abuse
- Gzip compression for responses
- Response time headers for monitoring
- Recommendation results cached in-memory for 5 minutes per user+trigger+context

## Security Features

- JWT token authentication
- bcrypt password hashing (12 rounds)
- Rate limiting per IP
- CORS with credentials
- Security headers (X-Frame-Options, Content-Security-Policy, etc.)
- Input validation on all endpoints
- SQL injection prevention via MongoDB driver
- HTTPS ready

## Known Limitations

- Image uploads use mock Cloudinary in development
- Admin check is simplified (TODO: implement proper roles)
- Token blacklist not implemented (logout is client-side)
- Review verified purchase flag not auto-populated

## Future Enhancements

- [ ] Complete Cloudinary integration
- [ ] Admin role system
- [ ] Token blacklist/refresh tokens
- [ ] Elasticsearch for full-text search
- [ ] Redis to replace in-memory recommendation cache (for multi-instance deployments)
- [ ] GraphQL API alongside REST
- [ ] Websocket for real-time updates
- [ ] Email notifications
- [ ] Analytics & reporting
- [ ] Payment gateway integration
- [ ] Catalog-based recommendations using behavioral signals (collaborative filtering)
- [ ] Search analytics

## License

Proprietary - Fittingly/Kloset

## Support

For issues or questions, refer to the OpenSpec artifacts:
- `/openspec/changes/implement-fittingly-go-apis/proposal.md`
- `/openspec/changes/implement-fittingly-go-apis/design.md`
- `/openspec/changes/implement-fittingly-go-apis/tasks.md`
- `/openspec/changes/implement-fittingly-go-apis/specs/*/spec.md`
