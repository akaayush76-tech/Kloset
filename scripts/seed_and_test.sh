#!/usr/bin/env bash
# Seed 12 wardrobe items and test the recommendation endpoint.
# Prerequisites: docker compose up -d (wait for server to be healthy on :8080)
# Safe to re-run: skips registration if user exists, skips seeding if items already present.

set -euo pipefail

BASE="http://localhost:8080/api"
SEP="────────────────────────────────────────────────"

echo "$SEP"
echo "  Kloset Recommendation Engine — Seed & Test"
echo "$SEP"

# ── 0. Rename wardrobe_items → wardrobeItems if the old collection still exists ──
echo ""
echo "▶  0/6  Checking collection name alignment..."
docker exec kloset-mongodb mongosh \
  "mongodb://root:rootpassword@localhost:27017/kloset_dev?authSource=admin" \
  --quiet --eval '
    const d = db.getSiblingDB("kloset_dev");
    const cols = d.listCollectionNames();
    if (cols.includes("wardrobe_items") && !cols.includes("wardrobeItems")) {
      d.wardrobe_items.renameCollection("wardrobeItems");
      print("Renamed wardrobe_items -> wardrobeItems");
    } else {
      print("Collection name OK");
    }
  ' 2>/dev/null \
  || echo "    (skipped — mongosh not in container PATH)"

# ── 1. Register test user (ignore conflict if already exists) ─────────────────
echo ""
echo "▶  1/6  Registering test user..."
curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name":     "Recs Tester",
    "email":    "recs@kloset.test",
    "password": "Test1234",
    "phone":    "9876543210",
    "gender":   "male"
  }' | python3 -m json.tool 2>/dev/null

# ── 2. Login and capture JWT ──────────────────────────────────────────────────
echo ""
echo "▶  2/6  Logging in..."
LOGIN=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"recs@kloset.test","password":"Test1234"}')
echo "$LOGIN" | python3 -m json.tool 2>/dev/null

TOKEN=$(echo "$LOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)
if [ -z "$TOKEN" ]; then
  echo "❌  Could not extract token. Is the server running on :8080?"
  exit 1
fi
echo ""
echo "✅  Token: ${TOKEN:0:50}..."

# ── 3. Seed 12 wardrobe items (skip if already seeded) ───────────────────────
echo ""
echo "▶  3/6  Seeding wardrobe items..."
echo "$SEP"

# post_item: all display output → stderr; only the bare ID → stdout
post_item() {
  local label="$1"
  local body="$2"
  echo -n "    ➕ $label ... " >&2
  local resp
  resp=$(curl -s -X POST "$BASE/wardrobe" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "$body")
  local id
  id=$(echo "$resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)
  echo "id=$id" >&2
  printf '%s' "$id"   # stdout only — no trailing newline to keep variable clean
}

EXISTING=$(curl -s "$BASE/wardrobe?limit=1" -H "Authorization: Bearer $TOKEN" \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['pagination']['total'])" 2>/dev/null || echo "0")

if [ "$EXISTING" -ge 12 ]; then
  echo "    Already seeded ($EXISTING items found) — skipping."
else
  post_item "White slim T-shirt (upper)" '{
    "name":"White Slim Tee","category":"upper","subtype":"t-shirt",
    "color":"white","image":"https://placeholder.com/white-tee.jpg",
    "brand":"Basics Co","size":"M","condition":"like-new","price":499,
    "identifiers":{"colorPrimary":"white","colorTone":"neutral","fit":"slim",
      "occasion":"casual","season":"all","formality":"casual","style":"minimalist","pattern":"solid"}
  }' > /dev/null

  post_item "Navy slim chinos (lower)" '{
    "name":"Navy Slim Chinos","category":"lower","subtype":"chinos",
    "color":"navy","image":"https://placeholder.com/navy-chinos.jpg",
    "brand":"CleanCut","size":"32","condition":"good","price":1299,
    "identifiers":{"colorPrimary":"navy","colorTone":"neutral","fit":"slim",
      "occasion":"smart_casual","season":"all","formality":"smart_casual","style":"classic","pattern":"solid"}
  }' > /dev/null

  post_item "Black oversized hoodie (outerwear)" '{
    "name":"Black Oversized Hoodie","category":"outerwear","subtype":"hoodie",
    "color":"black","image":"https://placeholder.com/black-hoodie.jpg",
    "brand":"UrbanLayer","size":"L","condition":"new","price":1799,
    "identifiers":{"colorPrimary":"black","colorTone":"neutral","fit":"oversized",
      "occasion":"casual","season":"winter","formality":"casual","style":"streetwear","pattern":"solid"}
  }' > /dev/null

  post_item "Dark blue slim jeans (lower)" '{
    "name":"Dark Blue Slim Jeans","category":"lower","subtype":"jeans",
    "color":"navy","image":"https://placeholder.com/dark-jeans.jpg",
    "brand":"DenimCo","size":"32","condition":"good","price":1499,
    "identifiers":{"colorPrimary":"navy","colorTone":"neutral","fit":"slim",
      "occasion":"casual","season":"all","formality":"casual","style":"classic","pattern":"solid"}
  }' > /dev/null

  post_item "White sneakers (shoes)" '{
    "name":"White Classic Sneakers","category":"shoes","subtype":"sneakers",
    "color":"white","image":"https://placeholder.com/white-sneakers.jpg",
    "brand":"SoleStep","size":"42","condition":"like-new","price":2499,
    "identifiers":{"colorPrimary":"white","colorTone":"neutral","fit":"regular",
      "occasion":"casual","season":"all","formality":"casual","style":"minimalist","pattern":"solid"}
  }' > /dev/null

  post_item "Grey regular sweatshirt (upper)" '{
    "name":"Grey Regular Sweatshirt","category":"upper","subtype":"sweatshirt",
    "color":"grey","image":"https://placeholder.com/grey-sweatshirt.jpg",
    "brand":"Cozy Basics","size":"L","condition":"good","price":899,
    "identifiers":{"colorPrimary":"grey","colorTone":"neutral","fit":"regular",
      "occasion":"weekend","season":"fall","formality":"casual","style":"minimalist","pattern":"solid"}
  }' > /dev/null

  post_item "Beige linen shirt (upper)" '{
    "name":"Beige Linen Shirt","category":"upper","subtype":"shirt",
    "color":"beige","image":"https://placeholder.com/beige-shirt.jpg",
    "brand":"SummerEdit","size":"M","condition":"new","price":1099,
    "identifiers":{"colorPrimary":"beige","colorTone":"earth","fit":"regular",
      "occasion":"smart_casual","season":"summer","formality":"smart_casual","style":"classic","pattern":"solid"}
  }' > /dev/null

  post_item "Black slim trousers (lower)" '{
    "name":"Black Slim Trousers","category":"lower","subtype":"trousers",
    "color":"black","image":"https://placeholder.com/black-trousers.jpg",
    "brand":"FitForm","size":"32","condition":"like-new","price":1699,
    "identifiers":{"colorPrimary":"black","colorTone":"neutral","fit":"slim",
      "occasion":"smart_casual","season":"all","formality":"smart_casual","style":"classic","pattern":"solid"}
  }' > /dev/null

  post_item "Brown leather loafers (shoes)" '{
    "name":"Brown Leather Loafers","category":"shoes","subtype":"loafers",
    "color":"brown","image":"https://placeholder.com/brown-loafers.jpg",
    "brand":"LeatherCraft","size":"42","condition":"good","price":3499,
    "identifiers":{"colorPrimary":"brown","colorTone":"earth","fit":"regular",
      "occasion":"smart_casual","season":"all","formality":"smart_casual","style":"classic","pattern":"solid"}
  }' > /dev/null

  post_item "Olive bomber jacket (outerwear)" '{
    "name":"Olive Bomber Jacket","category":"outerwear","subtype":"bomber",
    "color":"green","image":"https://placeholder.com/olive-bomber.jpg",
    "brand":"FieldEdit","size":"M","condition":"new","price":2999,
    "identifiers":{"colorPrimary":"green","colorTone":"earth","fit":"regular",
      "occasion":"casual","season":"fall","formality":"casual","style":"streetwear","pattern":"solid"}
  }' > /dev/null

  post_item "Striped casual shirt (upper)" '{
    "name":"Blue Stripe Casual Shirt","category":"upper","subtype":"shirt",
    "color":"blue","image":"https://placeholder.com/stripe-shirt.jpg",
    "brand":"PrintHouse","size":"M","condition":"good","price":799,
    "identifiers":{"colorPrimary":"blue","colorTone":"bold","fit":"regular",
      "occasion":"weekend","season":"spring","formality":"casual","style":"classic","pattern":"stripes"}
  }' > /dev/null

  post_item "Black leather belt (accessory)" '{
    "name":"Black Leather Belt","category":"accessory","subtype":"belt",
    "color":"black","image":"https://placeholder.com/black-belt.jpg",
    "brand":"BuckleUp","size":"M","condition":"new","price":599,
    "identifiers":{"colorPrimary":"black","colorTone":"neutral","fit":"regular",
      "occasion":"smart_casual","season":"all","formality":"smart_casual","style":"classic","pattern":"solid"}
  }' > /dev/null

  echo ""
  echo "✅  12 wardrobe items seeded."
fi

# ── 4. Fetch trigger ID from live wardrobe (last item = oldest = White Slim Tee)
echo ""
echo "▶  4/6  Fetching trigger item ID from wardrobe..."
echo "$SEP"
WARDROBE_RESP=$(curl -s "$BASE/wardrobe?limit=12" -H "Authorization: Bearer $TOKEN")
# Pick the item named "White Slim Tee"
TRIGGER_ID=$(echo "$WARDROBE_RESP" | python3 -c "
import sys, json
items = json.load(sys.stdin)['data']
match = next((i['id'] for i in items if i['name'] == 'White Slim Tee'), None)
if match:
    print(match)
" 2>/dev/null)

if [ -z "$TRIGGER_ID" ]; then
  echo "❌  Could not find 'White Slim Tee' in wardrobe. Seeding may have failed."
  exit 1
fi
echo "    Trigger ID (White Slim Tee): $TRIGGER_ID"

# ── 5. Test recommendations ───────────────────────────────────────────────────
echo ""
echo "▶  5/6  POST /api/recommendations/outfits — context: casual, include_shop_items: true"
echo "$SEP"

curl -s -X POST "$BASE/recommendations/outfits" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"trigger_item_id\":\"$TRIGGER_ID\",\"trigger_item_type\":\"closet\",\"context_filter\":\"casual\",\"limit\":5,\"include_shop_items\":true}" \
  | python3 -m json.tool 2>/dev/null

echo ""
echo "$SEP"
echo "▶  6/6  POST /api/recommendations/outfits — context: smart_casual, include_shop_items: false"
echo "$SEP"

curl -s -X POST "$BASE/recommendations/outfits" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"trigger_item_id\":\"$TRIGGER_ID\",\"trigger_item_type\":\"closet\",\"context_filter\":\"smart_casual\",\"limit\":5,\"include_shop_items\":false}" \
  | python3 -m json.tool 2>/dev/null

echo ""
echo "$SEP"
echo "  Done."
echo "$SEP"
