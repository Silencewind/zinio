# API Documentation

This document provides comprehensive documentation for all public APIs, functions, and components in the Zinio DRM removal tool.

## Table of Contents

- [Types](#types)
  - [MagazineID](#magazineid)
  - [IssueID](#issueid)
  - [Session](#session)
  - [Magazine](#magazine)
  - [IssueMetadata](#issuemetadata)
  - [Issue](#issue)
- [Functions](#functions)
  - [Login](#login)
  - [GetMagazines](#getmagazines)
  - [GetIssue](#getissue)
  - [GetURL](#geturl)

---

## Types

### MagazineID

`MagazineID` is a unique identifier of a magazine.

**Type:** `string`

**Example:**
```go
magazineID := MagazineID("12345")
```

---

### IssueID

`IssueID` is a unique identifier of an issue.

**Type:** `string`

**Example:**
```go
issueID := IssueID("67890")
```

---

### Session

`Session` contains session data for a single authenticated user. It is created after successful login and is required for all subsequent API calls.

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `login` | `string` | User's email address (private) |
| `password` | `string` | User's password (private) |
| `profileID` | `string` | User's profile ID obtained from authentication (private) |

**Note:** All fields are private. Use the `Login` function to create a `Session` instance.

**Example:**
```go
ctx := context.Background()
session, err := Login(ctx, "user@example.com", "password123")
if err != nil {
    log.Fatal(err)
}
// Use session for subsequent API calls
```

---

### Magazine

`Magazine` represents a single magazine with all its issues.

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `MagazineID` | Unique identifier of the magazine |
| `Title` | `string` | Display name/title of the magazine |
| `Issues` | `[]IssueMetadata` | List of all issues available for this magazine |

**Example:**
```go
magazines, err := session.GetMagazines(ctx)
if err != nil {
    log.Fatal(err)
}

for _, magazine := range magazines {
    fmt.Printf("Magazine: %s (ID: %s)\n", magazine.Title, magazine.ID)
    fmt.Printf("  Issues: %d\n", len(magazine.Issues))
    
    for _, issue := range magazine.Issues {
        fmt.Printf("    - %s (ID: %s)\n", issue.Title, issue.ID)
    }
}
```

---

### IssueMetadata

`IssueMetadata` contains ID and title of an issue. It is used within the `Magazine` struct to represent available issues.

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `IssueID` | Unique identifier of the issue |
| `Title` | `string` | Display name/title of the issue |

**Example:**
```go
magazines, err := session.GetMagazines(ctx)
if err != nil {
    log.Fatal(err)
}

// Access issue metadata from a magazine
for _, magazine := range magazines {
    for _, issueMetadata := range magazine.Issues {
        fmt.Printf("Issue: %s\n", issueMetadata.Title)
        fmt.Printf("Issue ID: %s\n", issueMetadata.ID)
    }
}
```

---

### Issue

`Issue` represents a single magazine issue with its metadata and download information.

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `IssueID` | Unique identifier of the issue |
| `Title` | `string` | Display name/title of the issue |
| `Password` | `string` | Decrypted password required to unlock the PDF pages |
| `PageCount` | `int` | Total number of pages in the issue |
| `baseURL` | `string` | Base URL for downloading pages (private) |

**Example:**
```go
magazineID := MagazineID("12345")
issueID := IssueID("67890")

issue, err := session.GetIssue(ctx, magazineID, issueID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Issue: %s\n", issue.Title)
fmt.Printf("Pages: %d\n", issue.PageCount)
fmt.Printf("Password: %s\n", issue.Password)

// Get URL for a specific page
for i := 0; i < issue.PageCount; i++ {
    url, err := issue.GetURL(i)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Page %d URL: %s\n", i, url)
}
```

---

## Functions

### Login

`Login` connects to the Zinio API server and authenticates a user using the given email and password.

**Signature:**
```go
func Login(ctx context.Context, email, password string) (*Session, error)
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `ctx` | `context.Context` | Context for controlling request cancellation and timeout |
| `email` | `string` | User's email address |
| `password` | `string` | User's password |

**Returns:**

- `*Session`: A session object containing authentication data, or `nil` on error
- `error`: An error if authentication fails

**Example:**
```go
package main

import (
    "context"
    "log"
)

func main() {
    ctx := context.Background()
    
    session, err := Login(ctx, "user@example.com", "mypassword")
    if err != nil {
        log.Fatalf("Failed to login: %v", err)
    }
    
    // Use session for subsequent API calls
    // ...
}
```

**Error Handling:**
```go
session, err := Login(ctx, email, password)
if err != nil {
    // Handle authentication errors
    // Common errors:
    // - Invalid credentials
    // - Network errors
    // - API service unavailable
    return fmt.Errorf("authentication failed: %w", err)
}
```

**Notes:**
- The function uses retry logic with exponential backoff for network resilience
- The context can be used to cancel the request or set a timeout
- The returned session is required for all subsequent API calls

---

### GetMagazines

`GetMagazines` downloads the list of all available magazines in the user's library.

**Signature:**
```go
func (session *Session) GetMagazines(ctx context.Context) ([]Magazine, error)
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `session` | `*Session` | Authenticated session (receiver) |
| `ctx` | `context.Context` | Context for controlling request cancellation and timeout |

**Returns:**

- `[]Magazine`: A slice of `Magazine` objects, each containing its ID, title, and list of issues
- `error`: An error if the request fails

**Example:**
```go
package main

import (
    "context"
    "fmt"
    "log"
)

func main() {
    ctx := context.Background()
    
    // First, authenticate
    session, err := Login(ctx, "user@example.com", "password")
    if err != nil {
        log.Fatal(err)
    }
    
    // Get all magazines
    magazines, err := session.GetMagazines(ctx)
    if err != nil {
        log.Fatalf("Failed to get magazines: %v", err)
    }
    
    fmt.Printf("Found %d magazines\n", len(magazines))
    
    for _, magazine := range magazines {
        fmt.Printf("\nMagazine: %s\n", magazine.Title)
        fmt.Printf("  ID: %s\n", magazine.ID)
        fmt.Printf("  Issues: %d\n", len(magazine.Issues))
    }
}
```

**Complete Example:**
```go
// List all magazines and their issues
magazines, err := session.GetMagazines(ctx)
if err != nil {
    return err
}

for _, magazine := range magazines {
    fmt.Printf("=== %s ===\n", magazine.Title)
    for i, issue := range magazine.Issues {
        fmt.Printf("%d. %s (ID: %s)\n", i+1, issue.Title, issue.ID)
    }
    fmt.Println()
}
```

**Error Handling:**
```go
magazines, err := session.GetMagazines(ctx)
if err != nil {
    // Handle errors:
    // - Session expired (need to re-authenticate)
    // - Network errors
    // - API service errors
    return fmt.Errorf("failed to retrieve magazines: %w", err)
}
```

**Notes:**
- Requires an authenticated session
- The function aggregates issues by magazine ID
- Each magazine contains all its available issues
- The function uses retry logic for network resilience

---

### GetIssue

`GetIssue` downloads metadata for a single magazine issue, including the decrypted password needed to unlock PDF pages.

**Signature:**
```go
func (session *Session) GetIssue(ctx context.Context, magazine MagazineID, issue IssueID) (*Issue, error)
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `session` | `*Session` | Authenticated session (receiver) |
| `ctx` | `context.Context` | Context for controlling request cancellation and timeout |
| `magazine` | `MagazineID` | Unique identifier of the magazine |
| `issue` | `IssueID` | Unique identifier of the issue |

**Returns:**

- `*Issue`: An `Issue` object containing ID, title, password, page count, and base URL, or `nil` on error
- `error`: An error if the request fails or decryption fails

**Example:**
```go
package main

import (
    "context"
    "fmt"
    "log"
)

func main() {
    ctx := context.Background()
    
    // Authenticate
    session, err := Login(ctx, "user@example.com", "password")
    if err != nil {
        log.Fatal(err)
    }
    
    // Get magazine list first
    magazines, err := session.GetMagazines(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get details for the first issue of the first magazine
    if len(magazines) > 0 && len(magazines[0].Issues) > 0 {
        magazineID := magazines[0].ID
        issueID := magazines[0].Issues[0].ID
        
        issue, err := session.GetIssue(ctx, magazineID, issueID)
        if err != nil {
            log.Fatalf("Failed to get issue: %v", err)
        }
        
        fmt.Printf("Issue: %s\n", issue.Title)
        fmt.Printf("Pages: %d\n", issue.PageCount)
        fmt.Printf("Password: %s\n", issue.Password)
    }
}
```

**Complete Workflow Example:**
```go
// Get all magazines
magazines, err := session.GetMagazines(ctx)
if err != nil {
    return err
}

// Process each magazine
for _, magazine := range magazines {
    fmt.Printf("Processing magazine: %s\n", magazine.Title)
    
    // Get details for each issue
    for _, issueMetadata := range magazine.Issues {
        issue, err := session.GetIssue(ctx, magazine.ID, issueMetadata.ID)
        if err != nil {
            log.Printf("Failed to get issue %s: %v", issueMetadata.Title, err)
            continue
        }
        
        fmt.Printf("  Issue: %s\n", issue.Title)
        fmt.Printf("    Pages: %d\n", issue.PageCount)
        
        // Download pages using issue.GetURL()
        for i := 0; i < issue.PageCount; i++ {
            url, err := issue.GetURL(i)
            if err != nil {
                log.Printf("Failed to get URL for page %d: %v", i, err)
                continue
            }
            fmt.Printf("    Page %d: %s\n", i, url)
        }
    }
}
```

**Error Handling:**
```go
issue, err := session.GetIssue(ctx, magazineID, issueID)
if err != nil {
    // Handle errors:
    // - Invalid magazine/issue ID
    // - Session expired
    // - Password decryption failure
    // - Network errors
    return fmt.Errorf("failed to retrieve issue: %w", err)
}
```

**Notes:**
- Requires an authenticated session
- The password field is automatically decrypted from the API response
- The password is required to unlock PDF pages when downloading
- The function uses retry logic for network resilience

---

### GetURL

`GetURL` returns the download URL for a specific page of an issue.

**Signature:**
```go
func (issue Issue) GetURL(page int) (string, error)
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `issue` | `Issue` | The issue object (receiver) |
| `page` | `int` | Zero-based page index (0 to PageCount-1) |

**Returns:**

- `string`: The full URL to download the PDF page
- `error`: An error if the page index is out of bounds

**Example:**
```go
package main

import (
    "context"
    "fmt"
    "log"
)

func main() {
    ctx := context.Background()
    
    // Authenticate and get issue
    session, err := Login(ctx, "user@example.com", "password")
    if err != nil {
        log.Fatal(err)
    }
    
    magazineID := MagazineID("12345")
    issueID := IssueID("67890")
    
    issue, err := session.GetIssue(ctx, magazineID, issueID)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get URLs for all pages
    for i := 0; i < issue.PageCount; i++ {
        url, err := issue.GetURL(i)
        if err != nil {
            log.Fatalf("Failed to get URL for page %d: %v", i, err)
        }
        fmt.Printf("Page %d: %s\n", i, url)
    }
}
```

**Error Handling:**
```go
url, err := issue.GetURL(pageIndex)
if err != nil {
    // Handle errors:
    // - Page index out of bounds (negative or >= PageCount)
    if err != nil {
        return fmt.Errorf("invalid page index %d: %w", pageIndex, err)
    }
}
```

**Validation:**
```go
// Validate page index before calling
if pageIndex < 0 || pageIndex >= issue.PageCount {
    return fmt.Errorf("page index %d is out of bounds (0-%d)", 
        pageIndex, issue.PageCount-1)
}

url, err := issue.GetURL(pageIndex)
if err != nil {
    return err
}
```

**Notes:**
- Page indices are zero-based (first page is 0, last page is PageCount-1)
- The function validates the page index and returns an error if out of bounds
- The returned URL points to a PDF file that requires the issue password to decrypt

---

## Complete Usage Example

Here's a complete example demonstrating how to use all the public APIs together:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
)

func main() {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    // Step 1: Authenticate
    email := "user@example.com"
    password := "password123"
    
    session, err := Login(ctx, email, password)
    if err != nil {
        log.Fatalf("Authentication failed: %v", err)
    }
    fmt.Println("✓ Successfully authenticated")
    
    // Step 2: Get all magazines
    magazines, err := session.GetMagazines(ctx)
    if err != nil {
        log.Fatalf("Failed to get magazines: %v", err)
    }
    fmt.Printf("✓ Found %d magazines\n", len(magazines))
    
    // Step 3: Process each magazine
    for _, magazine := range magazines {
        fmt.Printf("\n=== %s ===\n", magazine.Title)
        fmt.Printf("Magazine ID: %s\n", magazine.ID)
        fmt.Printf("Total Issues: %d\n\n", len(magazine.Issues))
        
        // Step 4: Get details for each issue
        for i, issueMetadata := range magazine.Issues {
            fmt.Printf("Issue %d: %s\n", i+1, issueMetadata.Title)
            
            issue, err := session.GetIssue(ctx, magazine.ID, issueMetadata.ID)
            if err != nil {
                log.Printf("  ✗ Failed to get issue details: %v", err)
                continue
            }
            
            fmt.Printf("  ✓ Pages: %d\n", issue.PageCount)
            fmt.Printf("  ✓ Password: %s\n", issue.Password)
            
            // Step 5: Get URLs for first few pages as example
            maxPages := 3
            if issue.PageCount < maxPages {
                maxPages = issue.PageCount
            }
            
            for j := 0; j < maxPages; j++ {
                url, err := issue.GetURL(j)
                if err != nil {
                    log.Printf("  ✗ Failed to get URL for page %d: %v", j, err)
                    continue
                }
                fmt.Printf("  Page %d URL: %s\n", j, url)
            }
            
            if issue.PageCount > maxPages {
                fmt.Printf("  ... and %d more pages\n", issue.PageCount-maxPages)
            }
            
            fmt.Println()
        }
    }
    
    fmt.Println("✓ Complete!")
}
```

---

## Error Handling Best Practices

### Context Usage

Always use context for cancellation and timeout control:

```go
// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

session, err := Login(ctx, email, password)
if err != nil {
    // Check if error is due to timeout
    if ctx.Err() == context.DeadlineExceeded {
        return fmt.Errorf("request timed out")
    }
    return err
}
```

### Retry Logic

The API functions include built-in retry logic, but you can implement additional retry strategies:

```go
var session *Session
var err error

maxRetries := 3
for i := 0; i < maxRetries; i++ {
    session, err = Login(ctx, email, password)
    if err == nil {
        break
    }
    time.Sleep(time.Second * time.Duration(i+1))
}

if err != nil {
    return fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}
```

### Error Wrapping

Wrap errors with context for better error messages:

```go
magazines, err := session.GetMagazines(ctx)
if err != nil {
    return fmt.Errorf("failed to retrieve magazines for user %s: %w", 
        session.login, err)
}
```

---

## Constants

The following constants are used internally but may be useful to know:

- `baseURL`: `"https://services.zinio.com/newsstandServices/"`
- `httpTimeout`: `5 * time.Second`
- `requestTimeout`: `time.Minute`

---

## Thread Safety

- `Session` instances are not thread-safe. Each goroutine should use its own session or implement proper synchronization.
- `Issue` instances are read-only after creation and can be safely used across goroutines.

---

## Notes

- All API functions require network connectivity to the Zinio service
- Authentication tokens are embedded in the session and are not exposed
- PDF passwords are automatically decrypted from encrypted API responses
- The API uses XML-based communication with the Zinio service
- All functions include automatic retry logic with exponential backoff for resilience
