# User Authentication API Documentation

This document outlines the authentication system for the Vinyl Record Catalog application.

## Key Components

### 1. User Registration

Allows new users to create an account with email, password, and name. The system verifies email uniqueness and securely hashes passwords before storage.

**Endpoint:** `POST /api/auth/register`

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securepassword",
  "name": "John Doe"
}
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2025-04-04T10:00:00Z",
    "updated_at": "2025-04-04T10:00:00Z"
  }
}
```

**Example Curl Command:**

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword",
    "name": "John Doe"
  }'
```

### 2. User Login

Authenticates users and provides a JWT token for subsequent API calls.

**Endpoint:** `POST /api/auth/login`

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2025-04-04T10:00:00Z",
    "updated_at": "2025-04-04T10:00:00Z"
  }
}
```

**Example Curl Command:**

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword"
  }'
```

### 3. Get User Profile

Retrieves the current user's profile information.

**Endpoint:** `GET /api/users/me`

**Headers:**

- `Authorization: Bearer {token}`

**Response:**

```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2025-04-04T10:00:00Z",
  "updated_at": "2025-04-04T10:00:00Z"
}
```

**Example Curl Command:**

```bash
curl -X GET http://localhost:8080/api/users/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 4. Update User Profile

Updates the user's profile information (email and name).

**Endpoint:** `PUT /api/users/me`

**Headers:**

- `Authorization: Bearer {token}`

**Request Body:**

```json
{
  "email": "updated@example.com",
  "name": "Updated Name"
}
```

**Response:**

```json
{
  "id": 1,
  "email": "updated@example.com",
  "name": "Updated Name",
  "created_at": "2025-04-04T10:00:00Z",
  "updated_at": "2025-04-04T10:30:00Z"
}
```

**Example Curl Command:**

```bash
curl -X PUT http://localhost:8080/api/users/me \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "email": "updated@example.com",
    "name": "Updated Name"
  }'
```

### 5. Update Password

Changes the user's password after verifying the current password.

**Endpoint:** `PUT /api/users/password`

**Headers:**

- `Authorization: Bearer {token}`

**Request Body:**

```json
{
  "current_password": "securepassword",
  "new_password": "newSecurePassword"
}
```

**Response:**

```json
{
  "message": "Password updated successfully"
}
```

**Example Curl Command:**

```bash
curl -X PUT http://localhost:8080/api/users/password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "current_password": "securepassword",
    "new_password": "newSecurePassword"
  }'
```

## Testing the Complete Authentication Flow

Follow these steps to test the entire authentication flow:

1. **Register a new user:**

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }' | jq -r '.token')

echo "Your token is: $TOKEN"
```

or 

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword"
  }' | jq -r '.token')

echo "Your token is: $TOKEN"
```

2. **Use the token to access protected resources:**

```bash
# List your vinyl records
curl -X GET http://localhost:8080/api/records \
  -H "Authorization: Bearer $TOKEN"

# Create a new record
curl -X POST http://localhost:8080/api/records \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Kind of Blue",
    "release_year": 1959,
    "catalog_number": "CL 1355",
    "condition": "Near Mint",
    "storage_location": "Shelf A",
    "artists": [{"id": 1, "role": "Primary Artist"}],
    "genres": [{"id": 1}]
  }'
```

3. **Update your profile:**

```bash
curl -X PUT http://localhost:8080/api/users/me \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "email": "updated@example.com",
    "name": "Updated Test User"
  }'
```

4. **Change your password:**

```bash
curl -X PUT http://localhost:8080/api/users/password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "current_password": "password123",
    "new_password": "newPassword123"
  }'
```

5. **Login with updated credentials:**

```bash
NEW_TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "updated@example.com",
    "password": "newPassword123"
  }' | jq -r '.token')

echo "Your new token is: $NEW_TOKEN"
```

## Security Considerations

- All passwords are hashed using bcrypt before storage
- JWT tokens are signed with a secret key
- Email uniqueness is enforced to prevent duplicate accounts
- Password verification occurs before changing passwords
- Error messages are designed to not leak sensitive information

## To get ip address of psql container on codespaces

```bash
DB_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' devcontainer-db-1)
```
