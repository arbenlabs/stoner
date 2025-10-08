# Stoner - Go API Utilities Library

A comprehensive collection of Go utilities designed to streamline API development and common backend operations. This library provides essential tools for validation, database operations, HTTP clients, logging, security, and more.

## Table of Contents

- [Installation](#installation)
- [Modules Overview](#modules-overview)
- [Usage Examples](#usage-examples)
  - [Assert Package](#assert-package)
  - [Crypto Package](#crypto-package)
  - [Database Package](#database-package)
  - [GQ Package](#gq-package)
  - [HTTP Package](#http-package)
  - [Logger Package](#logger-package)
  - [Middleware Package](#middleware-package)
  - [Sanitize Package](#sanitize-package)
  - [Time Package](#time-package)
  - [UUID Package](#uuid-package)
- [Contributing](#contributing)
- [License](#license)

## Installation

```bash
go get github.com/arbenlabs/stoner
```

## Modules Overview

| Package | Description | Key Features |
|---------|-------------|--------------|
| `assert` | Data validation and assertions | Type-safe validation, range checks, format validation |
| `crypto` | Cryptographic operations | Password hashing, AES encryption, HMAC signing |
| `db` | Database utilities | Connection management, query builder, migrations |
| `gq` | GORM query utilities | Generic CRUD operations, pagination, filtering |
| `http` | HTTP client utilities | Retry logic, circuit breaker, rate limiting |
| `logger` | Structured logging | JSON logging, context support, performance metrics |
| `middleware` | HTTP middleware | Rate limiting, CSRF protection, request validation |
| `sanitize` | Input sanitization | HTML/SQL sanitization, filename cleaning |
| `time` | Time utilities | Timezone handling, date calculations, cron scheduling |
| `uuid` | UUID generation | UUID v4 generation, validation, parsing |

## Usage Examples

### Assert Package

The `assert` package provides comprehensive data validation functions for various data types and formats.

```go
package main

import (
    "fmt"
    "github.com/arbenlabs/stoner/assert"
)

func main() {
    // String validation
    if err := assert.AssertNonEmptyString("hello"); err != nil {
        fmt.Println("Error:", err)
    }
    
    // Email validation
    if err := assert.AssertValidEmail("user@example.com"); err != nil {
        fmt.Println("Invalid email:", err)
    }
    
    // Range validation
    if err := assert.AssertInRange(5.0, 1.0, 10.0); err != nil {
        fmt.Println("Out of range:", err)
    }
    
    // Length validation
    if err := assert.AssertMinLength("password", 8); err != nil {
        fmt.Println("Too short:", err)
    }
    
    // Collection validation
    slice := []interface{}{1, 2, 3, 4, 5}
    if err := assert.AssertUnique(slice); err != nil {
        fmt.Println("Contains duplicates:", err)
    }
}
```

### Crypto Package

The `crypto` package provides secure cryptographic operations including password hashing, encryption, and digital signatures.

```go
package main

import (
    "fmt"
    "github.com/arbenlabs/stoner/crypto"
)

func main() {
    // Password hashing
    password := "mySecurePassword"
    hash, err := crypto.HashPassword(password)
    if err != nil {
        panic(err)
    }
    
    // Verify password
    isValid := crypto.VerifyPassword(password, hash)
    fmt.Println("Password valid:", isValid)
    
    // AES encryption
    key := []byte("32-byte-long-key-for-AES-256-GCM!")
    data := []byte("sensitive data")
    
    encrypted, err := crypto.EncryptAES(key, data)
    if err != nil {
        panic(err)
    }
    
    // Decrypt
    decrypted, err := crypto.DecryptAES(key, encrypted)
    if err != nil {
        panic(err)
    }
    fmt.Println("Decrypted:", string(decrypted))
    
    // HMAC signing
    signature := crypto.SignHMAC(key, data)
    isValidSignature := crypto.VerifyHMAC(key, data, signature)
    fmt.Println("Signature valid:", isValidSignature)
    
    // Generate secure random strings
    apiKey, err := crypto.GenerateAPIKey()
    if err != nil {
        panic(err)
    }
    fmt.Println("API Key:", apiKey)
}
```

### Database Package

The `db` package provides database connection management, query building, and migration utilities.

```go
package main

import (
    "fmt"
    "time"
    "github.com/arbenlabs/stoner/db"
)

func main() {
    // Database configuration
    config := &db.Config{
        Host:         "localhost",
        Port:         5432,
        Database:     "mydb",
        Username:     "user",
        Password:     "password",
        SSLMode:      "disable",
        MaxOpenConns: 25,
        MaxIdleConns: 10,
        MaxLifetime:  time.Hour,
        MaxIdleTime:  time.Minute * 30,
    }
    
    // Create connection
    conn, err := db.NewConnection(config)
    if err != nil {
        panic(err)
    }
    defer conn.Close()
    
    // Query builder
    qb := db.NewQueryBuilder()
    query, args := qb.
        Select("id", "name", "email").
        From("users").
        Where("active = ?", true).
        OrderBy("created_at", "DESC").
        Limit(10).
        Build()
    
    fmt.Println("Query:", query)
    fmt.Println("Args:", args)
    
    // Transaction
    tx, err := conn.BeginTransaction()
    if err != nil {
        panic(err)
    }
    
    // Execute within transaction
    _, err = tx.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", "John", "john@example.com")
    if err != nil {
        tx.Rollback()
        panic(err)
    }
    
    err = tx.Commit()
    if err != nil {
        panic(err)
    }
}
```

### GQ Package

The `gq` package provides generic GORM query utilities with built-in validation and security features.

```go
package main

import (
    "fmt"
    "github.com/arbenlabs/stoner/gq"
    "gorm.io/gorm"
)

type User struct {
    ID    string `gorm:"primaryKey"`
    Name  string `gorm:"column:name"`
    Email string `gorm:"column:email"`
    Age   int    `gorm:"column:age"`
}

func main() {
    // Assuming you have a GORM database connection
    var db *gorm.DB
    
    // Insert a single record
    user := User{Name: "John Doe", Email: "john@example.com", Age: 30}
    createdUser, err := gq.InsertRecord(db, user)
    if err != nil {
        panic(err)
    }
    fmt.Println("Created user:", createdUser)
    
    // Batch insert
    users := []User{
        {Name: "Alice", Email: "alice@example.com", Age: 25},
        {Name: "Bob", Email: "bob@example.com", Age: 35},
    }
    err = gq.BatchInsert(db, users, 100)
    if err != nil {
        panic(err)
    }
    
    // Get paginated records
    records, totalPages, err := gq.GetAllRecords[User](db, 1, 10)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Found %d users, %d total pages\n", len(records), totalPages)
    
    // Get record by field
    userByEmail, err := gq.GetRecordByField[User](db, "email", "john@example.com")
    if err != nil {
        panic(err)
    }
    fmt.Println("User by email:", userByEmail)
    
    // Filtered pagination
    conditions := map[string]interface{}{
        "age": 25,
        "name": "Alice",
    }
    filteredUsers, totalPages, err := gq.GetFilteredPaginatedRecords[User](db, 1, 10, conditions)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Filtered users: %d\n", len(filteredUsers))
    
    // Update record
    updates := map[string]interface{}{
        "age": 31,
    }
    err = gq.UpdateRecordByID[User, map[string]interface{}](db, user.ID, updates)
    if err != nil {
        panic(err)
    }
}
```

### HTTP Package

The `http` package provides a robust HTTP client with retry logic, circuit breaker, and rate limiting.

```go
package main

import (
    "fmt"
    "time"
    "github.com/arbenlabs/stoner/http"
)

func main() {
    // Create HTTP client
    client := http.NewClient("https://api.example.com")
    
    // Configure retry settings
    client.SetRetryConfig(&http.RetryConfig{
        MaxRetries: 3,
        Delay:      time.Second,
        Backoff:    2.0,
    })
    
    // Configure circuit breaker
    client.SetCircuitBreaker(&http.CircuitBreaker{
        MaxFailures:  5,
        Timeout:      30 * time.Second,
        ResetTimeout: 60 * time.Second,
    })
    
    // Set default headers
    client.SetDefaultHeaders(map[string]string{
        "User-Agent": "MyApp/1.0",
        "Accept":     "application/json",
    })
    
    // GET request
    resp, err := client.Get("/users", nil)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Status: %d\n", resp.StatusCode)
    fmt.Printf("Body: %s\n", string(resp.Body))
    
    // POST request with JSON
    userData := map[string]string{
        "name":  "John Doe",
        "email": "john@example.com",
    }
    
    resp, err = client.Post("/users", userData, nil)
    if err != nil {
        panic(err)
    }
    
    // JSON request/response
    var result map[string]interface{}
    err = client.GetJSON("/users/123", &result, nil)
    if err != nil {
        panic(err)
    }
    fmt.Println("User data:", result)
    
    // Rate limiter
    limiter := http.NewRateLimiter(time.Second, 10) // 10 requests per second
    
    for i := 0; i < 15; i++ {
        limiter.Wait() // Wait for rate limit
        resp, err := client.Get("/api/data", nil)
        if err != nil {
            fmt.Printf("Request %d failed: %v\n", i+1, err)
        } else {
            fmt.Printf("Request %d succeeded\n", i+1)
        }
    }
}
```

### Logger Package

The `logger` package provides structured logging with context support and performance metrics.

```go
package main

import (
    "context"
    "time"
    "github.com/arbenlabs/stoner/logger"
)

func main() {
    // Create logger configuration
    config := logger.NewLoggerConfig(
        "my-service",    // service name
        true,            // add source
        "1.0.0",         // version
        "production",    // environment
    )
    
    // Initialize logger
    log, err := logger.NewLogger(config)
    if err != nil {
        panic(err)
    }
    
    // Basic logging
    log.Info("Application started", "port", 8080)
    log.Warn("Deprecated API endpoint used", "endpoint", "/old-api")
    log.Error("Database connection failed", "error", "connection timeout")
    
    // Contextual logging
    ctx := context.Background()
    ctx = logger.ContextWithTraceID(ctx, "trace-123")
    ctx = logger.ContextWithUserID(ctx, "user-456")
    
    logWithContext := log.WithContext(ctx)
    logWithContext.Info("User action performed", "action", "login")
    
    // Conditional logging
    log.InfoIf(true, "Conditional message", "condition", "met")
    log.ErrorIf(false, "This won't be logged")
    
    // Error with stack trace
    err = fmt.Errorf("something went wrong")
    log.ErrorWithStack("Operation failed", err, "operation", "data-processing")
    
    // Structured error logging
    errorDetails := logger.ErrorDetails{
        Code:    "VALIDATION_ERROR",
        Message: "Invalid input provided",
        Details: map[string]interface{}{
            "field":   "email",
            "value":   "invalid-email",
            "reason":  "malformed format",
        },
    }
    log.LogError(err, errorDetails, "user_id", "user-123")
    
    // Performance logging
    start := time.Now()
    time.Sleep(100 * time.Millisecond) // Simulate work
    log.LogPerformance("database_query", time.Since(start), map[string]interface{}{
        "table":        "users",
        "rows_returned": 150,
    })
    
    // HTTP request logging
    log.LogHTTPRequest("GET", "/api/users", "Mozilla/5.0", "192.168.1.1", "application/json")
    
    // Security event logging
    log.LogSecurityEvent("failed_login", "invalid_credentials", "medium", map[string]interface{}{
        "ip_address": "192.168.1.100",
        "user_agent": "Mozilla/5.0",
        "attempts":   3,
    })
}
```

### Middleware Package

The `middleware` package provides HTTP middleware for security, rate limiting, and request validation.

```go
package main

import (
    "net/http"
    "github.com/arbenlabs/stoner/middleware"
    "github.com/arbenlabs/stoner/logger"
)

func main() {
    // Initialize logger
    config := logger.NewLoggerConfig("api-server", true, "1.0.0", "production")
    log, err := logger.NewLogger(config)
    if err != nil {
        panic(err)
    }
    
    // Create middleware
    mw := middleware.NewMiddleware(
        100,    // rate limit: 100 requests per second
        200,    // burst: 200 requests
        1024*1024*10, // max request size: 10MB
        1024*1024,    // max header size: 1MB
        1024*1024*50, // max file upload: 50MB
        30,     // read timeout: 30 seconds
        30,     // write timeout: 30 seconds
    )
    
    // Set logger
    mw.logger = log
    
    // Your handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    })
    
    // Apply middleware
    protectedHandler := mw.RateLimit(handler)
    protectedHandler = mw.RequestSizeLimit()(protectedHandler)
    protectedHandler = mw.RequestTimeout()(protectedHandler)
    protectedHandler = mw.LogHTTRequest(protectedHandler)
    
    // CSRF protection
    authKey := []byte("32-byte-long-auth-key-for-csrf!")
    csrfHandler := mw.CSRFMiddleware(authKey, true)(protectedHandler)
    
    // Start server
    http.ListenAndServe(":8080", csrfHandler)
}
```

### Sanitize Package

The `sanitize` package provides comprehensive input sanitization for security and data integrity.

```go
package main

import (
    "fmt"
    "github.com/arbenlabs/stoner/sanitize"
)

func main() {
    // String normalization
    text := "  Hello   World  "
    normalized := sanitize.NormalizeWhitespace(text)
    fmt.Println("Normalized:", normalized) // "Hello World"
    
    // HTML sanitization
    htmlContent := "<script>alert('xss')</script><p>Safe content</p>"
    safeHTML := sanitize.CleanHTML(htmlContent)
    fmt.Println("Safe HTML:", safeHTML) // "Safe content"
    
    // SQL sanitization
    userInput := "'; DROP TABLE users; --"
    safeSQL := sanitize.SanitizeSQLInput(userInput)
    fmt.Println("Safe SQL:", safeSQL)
    
    // Email sanitization
    email := "  USER@EXAMPLE.COM  "
    cleanEmail := sanitize.SanitizeEmail(email)
    fmt.Println("Clean email:", cleanEmail) // "user@example.com"
    
    // URL sanitization
    url := "javascript:alert('xss')"
    safeURL := sanitize.SanitizeURL(url)
    fmt.Println("Safe URL:", safeURL) // "" (empty, invalid scheme)
    
    // Filename sanitization
    filename := "../../../etc/passwd"
    safeFilename := sanitize.SanitizeFilename(filename)
    fmt.Println("Safe filename:", safeFilename) // "etc_passwd"
    
    // Path sanitization
    path := "../../../sensitive/file.txt"
    safePath := sanitize.SanitizePath(path)
    fmt.Println("Safe path:", safePath) // "sensitive/file.txt"
    
    // Special character removal
    textWithSpecial := "Hello@#$%World!"
    alphanumeric := sanitize.KeepOnlyAlphanumeric(textWithSpecial)
    fmt.Println("Alphanumeric only:", alphanumeric) // "HelloWorld"
    
    // Control character removal
    textWithControl := "Hello\x00\x01World"
    cleanText := sanitize.RemoveControlChars(textWithControl)
    fmt.Println("Clean text:", cleanText) // "HelloWorld"
    
    // Comprehensive sanitization
    dirtyInput := "<script>alert('xss')</script>  Hello   World  "
    cleanInput := sanitize.SanitizeString(dirtyInput)
    fmt.Println("Fully sanitized:", cleanInput)
    
    // Display-specific sanitization
    displayText := sanitize.SanitizeForDisplay(dirtyInput)
    fmt.Println("Display safe:", displayText)
    
    // Storage-specific sanitization
    storageText := sanitize.SanitizeForStorage(dirtyInput)
    fmt.Println("Storage safe:", storageText)
}
```

### Time Package

The `time` package provides comprehensive time utilities including timezone handling, date calculations, and scheduling.

```go
package main

import (
    "fmt"
    "time"
    "github.com/arbenlabs/stoner/time"
)

func main() {
    // Timezone handling
    tz, err := time.NewTimeZone("America/New_York")
    if err != nil {
        panic(err)
    }
    
    now := tz.Now()
    fmt.Println("Current time in NYC:", tz.Format(now, "2006-01-02 15:04:05"))
    
    // Date utilities
    date := time.NewDate(2024, 12, 25)
    fmt.Println("Christmas date:", date.String())
    fmt.Println("Is valid:", date.IsValid())
    
    // Duration utilities
    duration := time.NewDuration(2 * time.Hour)
    doubled := duration.Multiply(2)
    fmt.Println("Doubled duration:", doubled.String())
    
    // Time calculations
    calc := time.NewTimeCalculator()
    now = time.Now()
    
    fmt.Println("Start of day:", calc.StartOfDay(now))
    fmt.Println("End of day:", calc.EndOfDay(now))
    fmt.Println("Start of week:", calc.StartOfWeek(now))
    fmt.Println("End of week:", calc.EndOfWeek(now))
    fmt.Println("Start of month:", calc.StartOfMonth(now))
    fmt.Println("End of month:", calc.EndOfMonth(now))
    
    // Date arithmetic
    tomorrow := calc.AddDays(now, 1)
    nextMonth := calc.AddMonths(now, 1)
    nextYear := calc.AddYears(now, 1)
    
    fmt.Println("Tomorrow:", tomorrow)
    fmt.Println("Next month:", nextMonth)
    fmt.Println("Next year:", nextYear)
    
    // Time differences
    daysBetween := calc.DaysBetween(now, tomorrow)
    hoursBetween := calc.HoursBetween(now, tomorrow)
    
    fmt.Printf("Days between: %d\n", daysBetween)
    fmt.Printf("Hours between: %d\n", hoursBetween)
    
    // Weekend/weekday checks
    fmt.Println("Is weekend:", calc.IsWeekend(now))
    fmt.Println("Is weekday:", calc.IsWeekday(now))
    
    // Time formatting
    formatter := time.NewFormatTime()
    
    fmt.Println("RFC3339:", formatter.RFC3339(now))
    fmt.Println("ISO8601:", formatter.ISO8601(now))
    fmt.Println("Date only:", formatter.DateOnly(now))
    fmt.Println("Time only:", formatter.TimeOnly(now))
    fmt.Println("Human readable:", formatter.HumanReadable(now))
    
    // Time parsing
    parser := time.NewParseTime()
    
    parsedTime, err := parser.FromString("2024-12-25T10:30:00Z")
    if err != nil {
        panic(err)
    }
    fmt.Println("Parsed time:", parsedTime)
    
    // Unix timestamp parsing
    unixTime := parser.FromUnix(1703505000)
    fmt.Println("Unix time:", unixTime)
    
    // Cron-like scheduling
    cron := time.NewCron()
    
    // Add a job that runs every 5 minutes
    err = cron.AddJob("cleanup", "* * * * */5", func() {
        fmt.Println("Running cleanup job at", time.Now())
    })
    if err != nil {
        panic(err)
    }
    
    // Start the cron scheduler
    cron.Start()
    
    // Keep the program running
    time.Sleep(10 * time.Minute)
}
```

### UUID Package

The `uuid` package provides UUID generation, validation, and parsing utilities.

```go
package main

import (
    "fmt"
    "github.com/arbenlabs/stoner/uuid"
)

func main() {
    // Generate a new UUID v4
    newUUID, err := uuid.NewV4()
    if err != nil {
        panic(err)
    }
    fmt.Println("Generated UUID:", newUUID.String())
    
    // Generate UUID string directly
    uuidString, err := uuid.NewUUIDString()
    if err != nil {
        panic(err)
    }
    fmt.Println("UUID string:", uuidString)
    
    // Generate UUID with namespace
    namespacedUUID, err := uuid.NewWithNamespace("user")
    if err != nil {
        panic(err)
    }
    fmt.Println("Namespaced UUID:", namespacedUUID)
    
    // Must generate (panics on error)
    mustUUID := uuid.MustNewUUIDString()
    fmt.Println("Must UUID:", mustUUID)
    
    // Parse UUID from string
    parsedUUID, err := uuid.Parse(uuidString)
    if err != nil {
        panic(err)
    }
    fmt.Println("Parsed UUID:", parsedUUID.String())
    
    // Validate UUID format
    isValid := uuid.IsValid(uuidString)
    fmt.Println("Is valid UUID:", isValid)
    
    isValid = uuid.IsValid("invalid-uuid")
    fmt.Println("Is invalid UUID valid:", isValid)
    
    // Must parse (panics on error)
    mustParsedUUID := uuid.MustNewV4()
    fmt.Println("Must parsed UUID:", mustParsedUUID.String())
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
