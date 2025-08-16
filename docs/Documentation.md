
# SalaryAdvanceService Documentation

## Overview
The **SalaryAdvanceService** is a RESTful API developed for the OneTap GoLang Developer Challenge to manage customer data, transactions, and ratings for a Salary Advance Loan system. It processes customer and transaction data from JSON files, calculates customer ratings, and supports secure user authentication with RS256 JWTs.  

The system is built using **Clean Architecture** for scalability, **GORM** with **PostgreSQL** for data persistence, and the **Gin framework** for high-performance API endpoints.  

This documentation details the API, rating logic, and architectural choices, ensuring clarity and alignment with the challenge requirements.

---

## System Architecture
The service follows **Clean Architecture** to ensure scalability, maintainability, and separation of concerns, with the following layers:

- **Domain Layer:** Defines entities (`Customer`, `Transaction`, `user`) and interfaces (`CustomerRepository`, `CustomerUseCase`, ...). Located in `internal/domain`.
- **Use Case Layer:** Implements business logic (e.g., `ImportCustomers`, `CalculateCustomerRating`) in `internal/usecases`.
- **Controller Layer:** Handles HTTP requests/responses using Gin. Located in `api/controllers`.
- **Repository Layer:** Manages database operations with GORM and PostgreSQL. Located in `internal/repository`.
- **Configuration:** Uses environment variables (`DB_DSN`, `JWT_PRIVATE_KEY`, `JWT_PUBLIC_KEY`) for flexibility.

---

## Technology Stack

- **Gin Framework:** Lightweight, high-performance HTTP framework for Go, enabling fast API endpoints with middleware for authentication and rate limiting.
- **GORM with PostgreSQL:** ORM for PostgreSQL supporting efficient queries, migrations, and data validation. PostgreSQL ensures robust, scalable storage.
- **RS256 JWT Authentication:** Uses RSA256 asymmetric encryption for secure JWTs, validated via public/private key pairs.
- **Clean Architecture:** Separates concerns to minimize dependencies, enabling easy extension (e.g., new rating metrics) and testing (e.g., mocking `CustomerRepository`).

---

## Why Clean Architecture?

- **Scalability:** Independent layers allow adding features without modifying existing code.
- **Testability:** Interfaces enable unit tests with mocks (e.g., `mocks.CustomerRepository` in `customer.usecase_test.go`).

---

## Customer Rating Calculation
The customer rating assesses eligibility for salary advance loans based on transaction data from `transactions.json`.  
The formula is implemented in `internal/usecases/customer.go` (`CalculateCustomerRating`):

```

rating = (0.3 * countScore + 0.3 * volumeScore + 0.2 * durationScore + 0.2 * stabilityScore) * 10.0

````

### Formula Components

| Component       | Weight | Description |
|-----------------|--------|-------------|
| **countScore**  | 30%    | Number of transactions, normalized as `min(len(transactions)/10.0, 1.0)`. Max 10 transactions. Reflects transaction frequency. |
| **volumeScore** | 30%    | Total outgoing transaction volume (sum of amounts where `fromAccount` matches customer’s `accountNo`), normalized as `min(totalVolume/10000.0, 1.0)`. Max $10,000. Indicates financial capacity. |
| **durationScore** | 20%  | Duration between first and last transaction in days, normalized as `min(durationDays/365.0, 1.0)`. Max 1 year. Measures account activity history. |
| **stabilityScore** | 20% | Balance stability, calculated as `max(1.0 - stdDev(balances)/10000.0, 0.0)`. Max standard deviation $10,000. Lower volatility indicates lower risk. |

### Implementation Details

**Logic:**
- Fetch customer by `customerId` and their transactions by `accountNo` using `CustomerRepository`.
- Compute `countScore`, `volumeScore`, `durationScore`, `stabilityScore`.
- Final rating: weighted sum, multiplied by 10, rounded to one decimal place, clamped to [1.0, 10.0].

**Edge Cases:**
- No transactions → rating = 1.0
- Single transaction → durationScore uses a minimum duration
- Negative standard deviation → stabilityScore clamped to ≥ 0.0

**Code Example:**
```go
countScore := math.Min(float64(len(transactions))/10.0, 1.0)
volumeScore := math.Min(totalVolume/10000.0, 1.0)
durationDays := lastDate.Sub(firstDate).Hours() / 24
durationScore := math.Min(durationDays/365.0, 1.0)
stabilityScore := math.Max(1.0-stdDev/10000.0, 0.0)
totalRating := (0.3*countScore + 0.3*volumeScore + 0.2*durationScore + 0.2*stabilityScore) * 10.0
totalRating = math.Round(totalRating*10) / 10
if totalRating < 1.0 {
    totalRating = 1.0
} else if totalRating > 10.0 {
    totalRating = 10.0
}
````

---

## Justification of Weights

The weights prioritize **transaction activity** for salary advance loans:

* **RFM Model:** Emphasizes Frequency (`countScore`) and Monetary (`volumeScore`) for customer value.
* **Credit Scoring:** VantageScore prioritizes payment history (proxy for `countScore`) and balances (`volumeScore`) over length of history (`durationScore`).
* **Alternative Weighting:** Equal weights or risk-oriented weights misalign with short-term loan sizing.

---

## Error Handling

### Account Number and Name Error Example:  Provided in the image


* Invalid records excluded from `valid_customers` and rating calculations.

### General Validation:

* Uses `github.com/go-playground/validator/v10` to validate fields.
* Handles `accountNo` as float64 or string, converts to string stripping leading zeros.
* Logs detailed errors for each record.

---

## API Endpoints

All endpoints require **Bearer Token authentication** using RS256 JWT unless otherwise noted. Hosted at `http://localhost:8080/api/v0`.

### 1. Import Customers

* **Method:** POST
* **Endpoint:** `/customers/import`
* **Headers:**

  * `Content-Type: multipart/form-data`
  * `Authorization: Bearer <token>`

**Request Body (form-data):**

* `file`: JSON array of customer objects

**Response (201 Created):**

```json
{
  "message": "Customers imported",
  "data": [...],
  "logs": [...]
}
```

* Validates `customerName` and `accountNo`
* Generates unique `customerId`
* Logs invalid records

### 2. Get Customer by ID

* **Method:** GET
* **Endpoint:** `/customers/{customerId}`
* **Headers:** `Authorization: Bearer <token>`

**Response (200 OK):**

```json
{
  "data": {...}
}
```

### 3. Get All Customers

* **Method:** GET
* **Endpoint:** `/customers`

**Response (200 OK):**

```json
{
  "data": [...]
}
```

### 4. Get All Transactions

* **Method:** GET
* **Endpoint:** `/transactions`

**Response (200 OK):**

```json
{
  "data": [...]
}
```

### 5. Login

* **Method:** POST
* **Endpoint:** `/user/auth/login`
* **Headers:** `Content-Type: application/json`

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "password",
  "role": "uploader"
}
```

**Response (200 OK):**

```json
{
  "access_token": "<jwt_token>",
  "refresh_token": "<refresh_token>"
}
```

### 6. Send Invite

* **Method:** POST
* **Endpoint:** `/user/invite/send`

**Request Body:**

```json
{
  "email": "user@example.com"
}
```

**Response (200 OK):**

```json
{
  "message": "Invitation sent"
}
```

### 7. Register User

* **Method:** POST
* **Endpoint:** `/user/auth/register`

**Request Body:**

```json
{
  "token": "<invite_token>",
  "password": "newPassword123!"
}
```

**Response (201 Created):**

```json
{
  "message": "User registered successfully"
}
```

### 8. Refresh Token

* **Method:** POST
* **Endpoint:** `/users/refresh`

**Request Body:**

```json
{
  "refresh_token": "<refresh_token>"
}
```

**Response (200 OK):**

```json
{
  "access_token": "<new_access_token>",
  "refresh_token": "<new_refresh_token>"
}
```

### 9. Import Transactions

* **Method:** POST
* **Endpoint:** `/customers/transactions/import?allowOverdraft=true`

**Request Body (form-data):**

* `file`: JSON array of transactions

**Response (201 Created):**

```json
{
  "message": "Transactions imported",
  "data": [...],
  "logs": [...]
}
```

### 10. Get Customer Rating

* **Method:** GET
* **Endpoint:** `/customers/{customer_id}/rating`

**Response (200 OK):**

```json
{
  "customer_id": "CUST-12345678",
  "rating": 8.5
}
```

* Returns `1.0` if no transactions exist
* Clamps rating to \[1.0, 10.0]

---

## Scalability and Maintenance

* Clean Architecture: Loose coupling, easy to extend
* GORM/PostgreSQL: Supports large datasets
* Gin Framework: Lightweight, fast
* Unit Tests: Covers edge cases in `customer_controller_test.go` and `rating_test.go`
* Logging: Detailed logs for debugging

---

## References

* [RFM Analysis](https://www.investopedia.com/terms/r/rfm-recency-frequency-monetary-value.asp)
* [Vintage Score](https://www.academybank.com/article/scoring-models-how-to-calculate-vantagescore)
* [FICO Scores](https://www.myfico.com/credit-education/whats-in-your-credit-score)
* [Customer Engagement Scores](https://desku.io/blogs/customer-engagement-score)

---

## Setup Instructions

### Prerequisites

* Go 1.22+
* PostgreSQL 15+
* Dependencies:

  * github.com/gin-gonic/gin
  * gorm.io/gorm
  * gorm.io/driver/postgres
  * github.com/go-playground/validator/v10
  * github.com/vektra/mockery/v2

### Configuration

```bash
export DB_DSN="host=localhost user=postgres password=secret dbname=salary_advance port=5432"
export JWT_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----..."
export JWT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----..."
```

### Run Migrations

```bash
go run cmd/migrate/main.go
```

### Run Server

```bash
go run cmd/server/main.go
```

### Test

```bash
go test -v ./internal/...
```

### Troubleshooting

* **Database Errors:** Verify PostgreSQL is running and `DB_DSN` is correct.
* **JWT Errors:** Ensure valid RS256 keys.
* **Test Failures:** Regenerate mocks:

```bash
mockery --name=CustomerRepository --dir=internal/domain --output=internal/usecases/mocks
```

* **Validation Errors:** Check JSON file formats.

### Next Steps

* Implement batch rating calculations
* Add caching (e.g., Redis) for GET `/customers/{id}/rating`
* Enhance logging with structured JSON for monitoring

```

---

