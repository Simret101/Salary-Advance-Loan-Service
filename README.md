
````markdown
# SalaryAdvanceService

## Overview

The SalaryAdvanceService is a Go-based RESTful API developed for the OneTap GoLang Developer Challenge (August 9, 2025).
It manages customer data, transactions, and ratings for a Salary Advance Loan system, processing JSON inputs (`customers.json`, `transactions.json`, `sample_customers.json`).

The service:

* Validates records.
* Generates synthetic transactions for inactive customers.
* Computes a rating (1–10) based on transaction metrics.

Built with **Clean Architecture**, GORM/PostgreSQL, Gin, and RS256 JWT, it ensures **scalability, security, and maintainability**.

---

## Features

* **Authentication & Authorization:** RS256 JWT, bcrypt passwords, admin and uploader roles, rate-limiting for login attempts.
* **Data Validation:** Validates sample customer data against `customers.json` and logs errors.
* **Transaction Processing:** Maps transactions, generates synthetic transactions, supports overdraft control.
* **Customer Rating:** Calculates rating based on transaction count, volume, duration, and balance stability.
* **Modular Design:** Clean Architecture with unit tests for authentication, validation, and rating logic.
* **Validation Output:** Produces JSON logs for verified/invalid records.

---

## Requirements

### Authentication & Authorization

* RS256 asymmetric with both public & private key JWT-based authentication.
* Passwords stored securely using bcrypt.
* Roles: admin (manage users, send invites), uploader (import data).
* Rate-limiting on  (429 Too Many Requests).
* Bearer token required for protected endpoints.

### Data Validation

* Reads `customers.json`, `transactions.json`, `sample_customers.json`.
* Validates accountNo and customerName, marks records as verified true/false.
* Produces JSON logs with errors and normalized data.
* Stores valid records in PostgreSQL (`valid_customers` table).

### Transaction Generation & Rating

* Maps transactions to customers by accountNo.
* Generates synthetic transactions for inactive customers.
* Calculates rating:

```text
rating = (0.3*countScore + 0.3*volumeScore + 0.2*durationScore + 0.2*stabilityScore) * 10.0
```

---

## Deliverables

* **Source Code:** Modular Go code in `internal/domain`, `internal/usecases`, `api/controllers`, `internal/repository`.
* **Unit Tests:** Covers authentication, validation, and rating logic.
* **Validation Output:** JSON logs for each record.
* **Documentation:** README and Documentation.md with setup, validation, rating, and security instructions.

---

## System Architecture

### Clean Architecture

* **Domain Layer:** Defines `Customer`, `Transaction`, `User`, and interfaces.
* **Use Case Layer:** Implements business logic (`ImportCustomers`, `ImportTransactions`, `CalculateCustomerRating`).
* **Controller Layer:** Handles HTTP requests with Gin.
* **Repository Layer:** Database operations with GORM/PostgreSQL.
* **Configuration:** Uses environment variables (`DB_DSN`, `JWT_PRIVATE_KEY`, `JWT_PUBLIC_KEY`).

### Technology Stack

* **Gin:** HTTP framework for REST APIs.
* **GORM/PostgreSQL:** ORM for database operations.
* **RS256 JWT:** Secure authentication.
* **Validator:** Input validation with `github.com/go-playground/validator/v10`.

### Why Clean Architecture?

* **Scalability:** Independent layers for adding features.
* **Testability:** Interfaces enable unit tests with mocks.
* **Maintainability:** Reduces technical debt.

---

## Data Validation

### Process

* Reads JSON, normalizes `accountNo`.
* Compares `customerName` and `accountNo` against `customers.json`.
* Checks for duplicates in `valid_customers`.
* Generates `customerId` (e.g., `CUST-12345678`).

### Error Handling

* Invalid account number or name mismatch.
* Logs errors in JSON format.

```json
{
  "record_index": 31,
  "verified": false,
  "errors": ["name or account number does not match existing records in customers table"],
  "attempted_name": "MOLA MARYE ASAHL",
  "attempted_account_no": "727473400983"
}
```

### Validation Output

* Valid records stored; invalid records logged.

```json
[
  {
    "record_index": 31,
    "verified": false,
    "errors": ["name or account number does not match existing records in customers table"],
    "attempted_name": "MOLA MARYE ASAHL",
    "attempted_account_no": "727473400983"
  },
  {
    "record_index": 1,
    "verified": true,
    "normalized_record": {
      "customerId": "CUST-12345678",
      "customerName": "MEHADI ALIYE MOHAMMED",
      "accountNo": "1050001035901"
    }
  }
]
```

---

## Transaction Generation & Rating

### Transaction Processing

* Maps transactions by `fromAccount` and `toAccount`.
* Validates `amount`, `date`, and account existence.
* Generates synthetic transactions for inactive customers.

### Rating Calculation

```go
countScore := math.Min(float64(len(transactions))/10.0, 1.0)
volumeScore := math.Min(totalVolume/10000.0, 1.0)
durationScore := math.Min(durationDays/365.0, 1.0)
stabilityScore := math.Max(1.0-stdDev/10000.0, 0.0)
totalRating := (0.3*countScore + 0.3*volumeScore + 0.2*durationScore + 0.2*stabilityScore) * 10.0
totalRating = math.Round(totalRating*10)/10
if totalRating < 1.0 { totalRating = 1.0 } else if totalRating > 10.0 { totalRating = 10.0 }
```

* **countScore (30%)**
* **volumeScore (30%)**
* **durationScore (20%)**
* **stabilityScore (20%)**

**Edge Cases:** No transactions → rating 1.0. Single transaction → minimum duration assumed.
---

## References for rating calculation formula

* [Investopedia RFM Analysis](https://www.investopedia.com/terms/r/rfm-recency-frequency-monetary-value.asp)
* [Voyado RFM](https://support.voyado.com/hc/en-us/articles/360013834800)
* [Academy Bank VantageScore](https://www.academybank.com/personal/resources/scoring-models-how-to-calculate-vantagescore)
* [MyFICO FICO Scores](https://www.myfico.com/credit-education/whats-in-your-credit-score)
* Baesens, B. (2016). *Credit Risk Analytics*. O’Reilly Media.
---

## Security Measures

* **RS256 JWT Authentication**: Tokens required for protected endpoints.
* **Password Storage**: bcrypt hashed.
* **Rate-Limiting**:  returns 429 on excessive attempts.
* **Input Validation**: Sanitizes accountNo.

---

## Setup Instructions

### Prerequisites

* Go 1.22+
* PostgreSQL 15+
* Dependencies:

```bash
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/go-playground/validator/v10
go get github.com/vektra/mockery/v2
```

### Configuration

```bash
createdb salary_advance
export DB_DSN="host=localhost user=postgres password=secret dbname=salary_advance port=5432"
export JWT_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----..."
export JWT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----..."
```

### Start Service

```bash
go run cmd/main.go
```

API available at [http://localhost:8080/api/v0](http://localhost:8080/api/v0)

### Testing

```bash
go test -v ./internal/usecases...
```

Generate mocks if needed:

```bash
mockery --name=CustomerRepository --dir=internal/domain --output=internal/usecases/mocks
```

---

## API Usage

### Login

**POST** `/user/auth/login`

```json
{
  "email": "user@example.com",
  "password": "newPassword123!",
  "role": "uploader"
}
```

Response: `200 OK` with `access_token` and `refresh_token`.

### Import Customers

**POST** `/customers/import`
Headers: `Authorization: Bearer <token>`
Body: JSON file (e.g., `sample_customers.json`)
Response: `201 Created` with imported data and logs.

### Import Transactions

**POST** `/customers/transactions/import?allowOverdraft=true`
Headers: `Authorization: Bearer <token>`
Body: JSON file (e.g., `transactions.json`)
Response: `201 Created` with imported data and logs.

### Get Customer Rating

**GET** `/customers/{customer_id}/rating`
Headers: `Authorization: Bearer <token>`
Response:

```json
{
  "customer_id": "CUST-12345678",
  "rating": 8.5
}
```

Other endpoints:

* `GET /customers`
* `GET /customers/{customerId}`
* `GET /transactions`
* `POST /user/invite/send` (admin only)
* `POST /user/auth/register`
* `POST /users/refresh`

---

## Troubleshooting

* **Database Errors:** Ensure PostgreSQL is running and `DB_DSN` is correct.
* **JWT Errors:** Verify RS256 keys.
* **Validation Errors:** Check JSON file formats.
* **Test Failures:** Regenerate mocks with `mockery`.
* **Rating Issues:** Check transaction data; synthetic transactions guarantee activity.


---
### POSTMAN DOCUMENATION LINK
* [Postman Documentation](https://documenter.getpostman.com/view/40134617/2sB3BHkow8)

---
