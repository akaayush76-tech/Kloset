// Initialize Kloset MongoDB database with collections and indexes
// This script runs automatically when MongoDB container starts

db = db.getSiblingDB('kloset_dev');

// Create users collection with indexes
db.createCollection('users');
db.users.createIndex({ email: 1 }, { unique: true });
db.users.createIndex({ createdAt: -1 });

// Create products collection with indexes
db.createCollection('products');
db.products.createIndex({ name: 'text', description: 'text', brand: 'text' });
db.products.createIndex({ category: 1, subcategory: 1 });
db.products.createIndex({ price: 1 });
db.products.createIndex({ rating: -1 });
db.products.createIndex({ createdAt: -1 });
db.products.createIndex({ brand: 1 });

// Create orders collection with indexes
db.createCollection('orders');
db.orders.createIndex({ userId: 1, createdAt: -1 });
db.orders.createIndex({ orderStatus: 1 });
db.orders.createIndex({ paymentStatus: 1 });
db.orders.createIndex({ trackingNumber: 1 });

// Create reviews collection with indexes
db.createCollection('reviews');
db.reviews.createIndex({ userId: 1, productId: 1 }, { unique: true });
db.reviews.createIndex({ productId: 1, createdAt: -1 });
db.reviews.createIndex({ rating: 1 });
db.reviews.createIndex({ helpful: -1 });

// Create wardrobe_items collection with indexes
db.createCollection('wardrobe_items');
db.wardrobe_items.createIndex({ userId: 1, category: 1 });
db.wardrobe_items.createIndex({ userId: 1, subtype: 1 });
db.wardrobe_items.createIndex({ userId: 1, createdAt: -1 });

// Insert sample data (optional)
// Users with sample data
db.users.insertOne({
  _id: new ObjectId(),
  name: "Demo User",
  email: "demo@example.com",
  password: "$2a$12$...", // bcrypt hash - will be replaced by actual user registration
  phone: "9876543210",
  gender: "male",
  dateOfBirth: new Date("1990-01-01"),
  avatar: "",
  preferences: {
    newsletter: true,
    notifications: true,
    favoriteColors: [],
    favoriteSizes: []
  },
  wishlist: [],
  cart: [],
  orders: [],
  reviews: [],
  wardrobeItems: [],
  isActive: true,
  createdAt: new Date(),
  updatedAt: new Date()
});

// Sample products
db.products.insertMany([
  {
    _id: new ObjectId(),
    name: "Classic Blue T-Shirt",
    description: "A comfortable and versatile blue t-shirt perfect for everyday wear",
    price: 799,
    originalPrice: 999,
    discount: 20,
    category: "upper",
    subcategory: "t-shirt",
    brand: "FitStyle",
    images: ["https://via.placeholder.com/300x400?text=Blue+Tshirt"],
    colors: ["blue", "navy", "light-blue"],
    sizes: ["XS", "S", "M", "L", "XL", "XXL"],
    rating: 4.5,
    reviewCount: 120,
    features: ["Cotton", "Breathable", "Easy Care"],
    fitTags: ["slim-fit", "casual"],
    combinations: ["jeans", "shorts"],
    outfits: [],
    deliveryDays: 3,
    isActive: true,
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    _id: new ObjectId(),
    name: "Dark Denim Jeans",
    description: "Premium dark denim jeans with perfect fit and comfort",
    price: 1499,
    originalPrice: 1999,
    discount: 25,
    category: "lower",
    subcategory: "jeans",
    brand: "DenimCo",
    images: ["https://via.placeholder.com/300x400?text=Denim+Jeans"],
    colors: ["dark-blue", "black"],
    sizes: ["28", "30", "32", "34", "36", "38"],
    rating: 4.7,
    reviewCount: 250,
    features: ["Cotton Blend", "Stretchable", "Durable"],
    fitTags: ["slim-fit", "regular-fit"],
    combinations: ["t-shirt", "shirt", "blazer"],
    outfits: [],
    deliveryDays: 3,
    isActive: true,
    createdAt: new Date(),
    updatedAt: new Date()
  }
]);

print("MongoDB initialization complete!");
print("Database: kloset_dev");
print("Collections created: users, products, orders, reviews, wardrobe_items");
print("Indexes created for all collections");
