package sanitize

import (
	"html"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// **************************************************
// Sanitize
// Sanitize is a package that provides functions to sanitize strings.
// **************************************************

// **************************************************
// --------------------------------------------------
// String Normalization Functions
// --------------------------------------------------
// **************************************************

// TrimAndLowercase trims whitespace and converts to lowercase
func TrimAndLowercase(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// TrimAndUppercase trims whitespace and converts to uppercase
func TrimAndUppercase(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

// TrimAndTitleCase trims whitespace and converts to title case
func TrimAndTitleCase(s string) string {
	return strings.Title(strings.TrimSpace(s))
}

// NormalizeWhitespace replaces multiple whitespace characters with single spaces
func NormalizeWhitespace(s string) string {
	// Replace multiple whitespace characters with single space
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

// RemoveExtraSpaces removes leading, trailing, and multiple consecutive spaces
func RemoveExtraSpaces(s string) string {
	return NormalizeWhitespace(s)
}

// TrimAll trims all specified characters from both ends
func TrimAll(s string, cutset string) string {
	return strings.Trim(s, cutset)
}

// **************************************************
// --------------------------------------------------
// HTML Sanitization Functions
// --------------------------------------------------
// **************************************************

// EscapeHTML escapes HTML special characters
func EscapeHTML(s string) string {
	return html.EscapeString(s)
}

// UnescapeHTML unescapes HTML entities
func UnescapeHTML(s string) string {
	return html.UnescapeString(s)
}

// RemoveHTMLTags removes all HTML tags from a string
func RemoveHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}

// StripHTMLComments removes HTML comments
func StripHTMLComments(s string) string {
	re := regexp.MustCompile(`<!--.*?-->`)
	return re.ReplaceAllString(s, "")
}

// CleanHTML removes HTML tags and comments, then escapes remaining content
func CleanHTML(s string) string {
	cleaned := RemoveHTMLTags(s)
	cleaned = StripHTMLComments(cleaned)
	return EscapeHTML(cleaned)
}

// **************************************************
// --------------------------------------------------
// SQL Sanitization Functions
// --------------------------------------------------
// **************************************************

// EscapeSQL escapes single quotes for SQL
func EscapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// RemoveSQLKeywords removes common SQL keywords that could be dangerous
func RemoveSQLKeywords(s string) string {
	sqlKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER",
		"EXEC", "EXECUTE", "UNION", "SCRIPT", "SCRIPT>", "<SCRIPT",
	}

	result := s
	for _, keyword := range sqlKeywords {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(keyword))
		result = re.ReplaceAllString(result, "")
	}
	return result
}

// SanitizeSQLInput sanitizes input for SQL queries
func SanitizeSQLInput(s string) string {
	// Remove SQL keywords
	sanitized := RemoveSQLKeywords(s)
	// Escape single quotes
	sanitized = EscapeSQL(sanitized)
	// Remove semicolons
	sanitized = strings.ReplaceAll(sanitized, ";", "")
	return sanitized
}

// **************************************************
// --------------------------------------------------
// Email and URL Sanitization
// --------------------------------------------------
// **************************************************

// SanitizeEmail cleans and normalizes email addresses
func SanitizeEmail(email string) string {
	// Trim and lowercase
	email = TrimAndLowercase(email)
	// Remove extra spaces
	email = NormalizeWhitespace(email)
	// Remove any characters that shouldn't be in email
	re := regexp.MustCompile(`[^\w@.-]`)
	email = re.ReplaceAllString(email, "")
	return email
}

// SanitizeURL cleans and validates URLs
func SanitizeURL(rawURL string) string {
	// Trim whitespace
	rawURL = strings.TrimSpace(rawURL)

	// Parse URL to validate format
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Only allow http and https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ""
	}

	// Return cleaned URL
	return parsedURL.String()
}

// **************************************************
// --------------------------------------------------
// Filename and Path Sanitization
// --------------------------------------------------
// **************************************************

// SanitizeFilename removes dangerous characters from filenames
func SanitizeFilename(filename string) string {
	// Remove path separators and other dangerous characters
	dangerousChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}

	for _, char := range dangerousChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	// Remove leading/trailing dots and spaces
	filename = strings.Trim(filename, ". ")

	// Limit length
	if len(filename) > 255 {
		filename = filename[:255]
	}

	return filename
}

// SanitizePath removes dangerous path components
func SanitizePath(path string) string {
	// Remove path traversal attempts
	path = strings.ReplaceAll(path, "../", "")
	path = strings.ReplaceAll(path, "..\\", "")
	path = strings.ReplaceAll(path, "..", "")

	// Remove null bytes
	path = strings.ReplaceAll(path, "\x00", "")

	// Normalize path separators
	path = strings.ReplaceAll(path, "\\", "/")

	return path
}

// **************************************************
// --------------------------------------------------
// Special Character and Encoding Functions
// --------------------------------------------------
// **************************************************

// RemoveSpecialChars removes all non-alphanumeric characters except spaces
func RemoveSpecialChars(s string) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// KeepOnlyAlphanumeric removes all characters except letters and numbers
func KeepOnlyAlphanumeric(s string) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// RemoveControlChars removes control characters (ASCII 0-31)
func RemoveControlChars(s string) string {
	var result strings.Builder
	for _, r := range s {
		if r >= 32 || r == '\n' || r == '\r' || r == '\t' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// NormalizeUnicode normalizes Unicode characters
func NormalizeUnicode(s string) string {
	// This is a simplified version - in production you might want to use golang.org/x/text/unicode/norm
	return strings.ToLower(s)
}

// RemoveEmojis removes emoji characters
func RemoveEmojis(s string) string {
	var result strings.Builder
	for _, r := range s {
		// Check if rune is an emoji (simplified check)
		if r < 0x1F600 || r > 0x1F64F { // Not in emoji range
			if r < 0x1F300 || r > 0x1F5FF { // Not in misc symbols range
				if r < 0x1F680 || r > 0x1F6FF { // Not in transport range
					if r < 0x1F700 || r > 0x1F77F { // Not in alchemical symbols
						if r < 0x1F780 || r > 0x1F7FF { // Not in geometric shapes extended
							if r < 0x1F800 || r > 0x1F8FF { // Not in supplemental arrows-C
								if r < 0x1F900 || r > 0x1F9FF { // Not in supplemental symbols and pictographs
									if r < 0x1FA00 || r > 0x1FA6F { // Not in chess symbols
										if r < 0x1FA70 || r > 0x1FAFF { // Not in symbols and pictographs extended-A
											result.WriteRune(r)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return result.String()
}

// **************************************************
// --------------------------------------------------
// General Purpose Sanitization
// --------------------------------------------------
// **************************************************

// SanitizeString performs comprehensive string sanitization
func SanitizeString(s string) string {
	// Remove control characters
	s = RemoveControlChars(s)
	// Normalize whitespace
	s = NormalizeWhitespace(s)
	// Trim spaces
	s = strings.TrimSpace(s)
	// Remove HTML tags
	s = RemoveHTMLTags(s)
	// Escape HTML
	s = EscapeHTML(s)
	return s
}

// SanitizeForDisplay sanitizes text for safe display
func SanitizeForDisplay(s string) string {
	// Remove HTML tags and comments
	s = RemoveHTMLTags(s)
	s = StripHTMLComments(s)
	// Escape HTML
	s = EscapeHTML(s)
	// Normalize whitespace
	s = NormalizeWhitespace(s)
	// Remove control characters
	s = RemoveControlChars(s)
	return strings.TrimSpace(s)
}

// SanitizeForStorage sanitizes text for safe storage
func SanitizeForStorage(s string) string {
	// Remove control characters
	s = RemoveControlChars(s)
	// Normalize whitespace
	s = NormalizeWhitespace(s)
	// Remove SQL keywords
	s = RemoveSQLKeywords(s)
	// Escape SQL
	s = EscapeSQL(s)
	return strings.TrimSpace(s)
}

// SanitizeForFilename sanitizes text for use as filename
func SanitizeForFilename(s string) string {
	// Remove HTML tags
	s = RemoveHTMLTags(s)
	// Remove special characters
	s = RemoveSpecialChars(s)
	// Normalize whitespace
	s = NormalizeWhitespace(s)
	// Replace spaces with underscores
	s = strings.ReplaceAll(s, " ", "_")
	// Sanitize filename
	s = SanitizeFilename(s)
	return s
}
