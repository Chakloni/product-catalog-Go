# Product Catalog API Documentation

-----

## Setup & Installation

### 1.- Clone the repository

Start by cloning the project and navigating into the directory:

```bash
git clone https://github.com/YOUR_USERNAME/product-catalog.git
cd product-catalog
```

### 2.- Initialize Go modules

Fetch the necessary dependencies:

```bash
go mod tidy
```

### 3.- Configure MongoDB Connection

You must configure the database connection string.

**Edit the file:**

```bash
internal/config/config.go
```

**Replace the placeholder** in the function `GetMongoURI()` with your **MongoDB Atlas connection string**.

**Example:**

```bash
return "mongodb-srv://<username>:<password>@cluster0.mongodb.net/"
```

-----

## Run the API

Execute the following command from the project root:

```bash
go run ./cmd/api
```

The server will be running at:

```bash
http://localhost:8080
```

-----

## API Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| **POST** | `/v1/products` | Create a **new product**. |
| **GET** | `/v1/products` | **List all non-deleted products** (supports pagination, filtering, sorting). |
| **GET** | `/v1/products/:id` | Get a **single product by ID**. |
| **PATCH** | `/v1/products/:id` | **Update** an existing product. |
| **DELETE** | `/v1/products/:id` | **Soft delete** a product (mark as deleted). |

-----

## Soft Delete Logic

Deleted products are **not physically removed** from the MongoDB database. Instead, they are marked as deleted by updating two fields in the document:

```json
{
  "is_deleted": true,
  "updated_at": "2025-10-29T18:25:43Z"
}
```

By default, all **GET** queries to the database **automatically exclude** products where `"is_deleted"` is `true`.

-----

## Example Product JSON

This structure is used when **creating** or **updating** a product:

```json
{
  "sku": "ABC-123",
  "name": "Wireless Mouse",
  "description": "Ergonomic mouse with 2.4GHz connection",
  "category": "Electronics",
  "price_cents": 2599,
  "currency": "USD",
  "stock": 25,
  "images": ["https://example.com/mouse.jpg"],
  "attributes": {"color": "black", "connectivity": "wireless"},
  "is_active": true
}
```

-----

## Useful Commands

### Format code

Ensure your Go code is properly formatted:

```bash
go fmt ./...
```

### Run tests (if added later)

Execute all tests in the project:

```bash
go test ./...
```

-----
