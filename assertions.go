package assertions

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// AssertNonEmptyString checks if a string is not empty
func AssertNonEmptyString(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("string cannot be empty")
	}
	return nil
}

// AssertNonZeroInt checks if an integer is not zero
func AssertNonZeroInt(value int) error {
	if value == 0 {
		return errors.New("integer cannot be zero")
	}
	return nil
}

// AssertNonZeroInt64 checks if an int64 is not zero
func AssertNonZeroInt64(value int64) error {
	if value == 0 {
		return errors.New("int64 cannot be zero")
	}
	return nil
}

// AssertNonZeroInt32 checks if an int32 is not zero
func AssertNonZeroInt32(value int32) error {
	if value == 0 {
		return errors.New("int32 cannot be zero")
	}
	return nil
}

// AssertNonZeroFloat64 checks if a float64 is not zero
func AssertNonZeroFloat64(value float64) error {
	if value == 0 {
		return errors.New("float64 cannot be zero")
	}
	return nil
}

// AssertNonZeroFloat32 checks if a float32 is not zero
func AssertNonZeroFloat32(value float32) error {
	if value == 0 {
		return errors.New("float32 cannot be zero")
	}
	return nil
}

// AssertPositiveInt checks if an integer is positive
func AssertPositiveInt(value int) error {
	if value <= 0 {
		return errors.New("integer must be positive")
	}
	return nil
}

// AssertPositiveInt64 checks if an int64 is positive
func AssertPositiveInt64(value int64) error {
	if value <= 0 {
		return errors.New("int64 must be positive")
	}
	return nil
}

// AssertPositiveInt32 checks if an int32 is positive
func AssertPositiveInt32(value int32) error {
	if value <= 0 {
		return errors.New("int32 must be positive")
	}
	return nil
}

// AssertNonEmptySlice checks if a slice is not empty
func AssertNonEmptySlice(value []any) error {
	if len(value) == 0 {
		return errors.New("slice cannot be empty")
	}
	return nil
}

// AssertNonEmptyMap checks if a map is not empty
func AssertNonEmptyMap(value map[any]any) error {
	if len(value) == 0 {
		return errors.New("map cannot be empty")
	}
	return nil
}

// AssertNonEmptyStruct checks if a struct is not empty
func AssertNonEmptyStruct(value any) error {
	if value == nil {
		return errors.New("struct cannot be empty")
	}
	return nil
}

// AssertNonEmptyInterface checks if an interface is not empty
func AssertNonEmptyInterface(value any) error {
	if value == nil {
		return errors.New("interface cannot be empty")
	}
	return nil
}

// AssertNonEmptyPointer checks if a pointer is not nil
func AssertNonEmptyPointer(value any) error {
	if value == nil {
		return errors.New("pointer cannot be nil")
	}
	return nil
}

// AssertNonEmptyTime checks if a time is not empty
func AssertNonEmptyTime(value time.Time) error {
	if value.IsZero() {
		return errors.New("time cannot be empty")
	}
	return nil
}

// Range/Boundary Assertions

// AssertInRange checks if a number is within a specific range (inclusive)
func AssertInRange(value, min, max float64) error {
	if value < min || value > max {
		return fmt.Errorf("value %v must be between %v and %v", value, min, max)
	}
	return nil
}

// AssertMinLength checks if a string or slice has minimum length
func AssertMinLength(value interface{}, minLength int) error {
	var length int
	switch v := value.(type) {
	case string:
		length = len(v)
	case []interface{}:
		length = len(v)
	case []string:
		length = len(v)
	case []int:
		length = len(v)
	case []int64:
		length = len(v)
	case []int32:
		length = len(v)
	case []float64:
		length = len(v)
	case []float32:
		length = len(v)
	default:
		return errors.New("unsupported type for length assertion")
	}

	if length < minLength {
		return fmt.Errorf("length %d must be at least %d", length, minLength)
	}
	return nil
}

// AssertMaxLength checks if a string or slice has maximum length
func AssertMaxLength(value interface{}, maxLength int) error {
	var length int
	switch v := value.(type) {
	case string:
		length = len(v)
	case []interface{}:
		length = len(v)
	case []string:
		length = len(v)
	case []int:
		length = len(v)
	case []int64:
		length = len(v)
	case []int32:
		length = len(v)
	case []float64:
		length = len(v)
	case []float32:
		length = len(v)
	default:
		return errors.New("unsupported type for length assertion")
	}

	if length > maxLength {
		return fmt.Errorf("length %d must be at most %d", length, maxLength)
	}
	return nil
}

// AssertMinValue checks if a numeric value is at least the minimum
func AssertMinValue(value, minValue float64) error {
	if value < minValue {
		return fmt.Errorf("value %v must be at least %v", value, minValue)
	}
	return nil
}

// AssertMaxValue checks if a numeric value is at most the maximum
func AssertMaxValue(value, maxValue float64) error {
	if value > maxValue {
		return fmt.Errorf("value %v must be at most %v", value, maxValue)
	}
	return nil
}

// Type-Specific Assertions

// AssertValidEmail checks if a string is a valid email format
func AssertValidEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}
	return nil
}

// AssertValidURL checks if a string is a valid URL format
func AssertValidURL(url string) error {
	urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`)
	if !urlRegex.MatchString(url) {
		return fmt.Errorf("invalid URL format: %s", url)
	}
	return nil
}

// AssertValidUUID checks if a string is a valid UUID format
func AssertValidUUID(uuid string) error {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(strings.ToLower(uuid)) {
		return fmt.Errorf("invalid UUID format: %s", uuid)
	}
	return nil
}

// AssertValidJSON checks if a string is valid JSON
func AssertValidJSON(jsonStr string) error {
	var js json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}
	return nil
}

// Collection Assertions

// AssertContains checks if a slice contains a specific value
func AssertContains(slice []interface{}, value interface{}) error {
	for _, item := range slice {
		if item == value {
			return nil
		}
	}
	return fmt.Errorf("slice does not contain value: %v", value)
}

// AssertUnique checks if a slice contains only unique values
func AssertUnique(slice []interface{}) error {
	seen := make(map[interface{}]bool)
	for _, item := range slice {
		if seen[item] {
			return fmt.Errorf("slice contains duplicate value: %v", item)
		}
		seen[item] = true
	}
	return nil
}

// AssertSorted checks if a slice is sorted in ascending order
func AssertSorted(slice []interface{}) error {
	if len(slice) <= 1 {
		return nil
	}

	for i := 1; i < len(slice); i++ {
		if !isLess(slice[i-1], slice[i]) {
			return fmt.Errorf("slice is not sorted at index %d", i)
		}
	}
	return nil
}

// Helper function to compare values for sorting
func isLess(a, b interface{}) bool {
	switch va := a.(type) {
	case int:
		if vb, ok := b.(int); ok {
			return va < vb
		}
	case int64:
		if vb, ok := b.(int64); ok {
			return va < vb
		}
	case int32:
		if vb, ok := b.(int32); ok {
			return va < vb
		}
	case float64:
		if vb, ok := b.(float64); ok {
			return va < vb
		}
	case float32:
		if vb, ok := b.(float32); ok {
			return va < vb
		}
	case string:
		if vb, ok := b.(string); ok {
			return va < vb
		}
	}
	return false
}

// Conditional Assertions

// AssertTrue checks if a boolean value is true
func AssertTrue(value bool) error {
	if !value {
		return errors.New("value must be true")
	}
	return nil
}

// AssertFalse checks if a boolean value is false
func AssertFalse(value bool) error {
	if value {
		return errors.New("value must be false")
	}
	return nil
}

// AssertEqual checks if two values are equal
func AssertEqual(actual, expected interface{}) error {
	if actual != expected {
		return fmt.Errorf("expected %v, got %v", expected, actual)
	}
	return nil
}

// AssertNotEqual checks if two values are not equal
func AssertNotEqual(actual, expected interface{}) error {
	if actual == expected {
		return fmt.Errorf("values should not be equal, both are %v", actual)
	}
	return nil
}

// AssertGreaterThan checks if the first value is greater than the second
func AssertGreaterThan(actual, expected float64) error {
	if actual <= expected {
		return fmt.Errorf("value %v must be greater than %v", actual, expected)
	}
	return nil
}

// AssertLessThan checks if the first value is less than the second
func AssertLessThan(actual, expected float64) error {
	if actual >= expected {
		return fmt.Errorf("value %v must be less than %v", actual, expected)
	}
	return nil
}

// Time Assertions

// AssertAfter checks if a time is after another time
func AssertAfter(actual, expected time.Time) error {
	if !actual.After(expected) {
		return fmt.Errorf("time %v must be after %v", actual, expected)
	}
	return nil
}

// AssertBefore checks if a time is before another time
func AssertBefore(actual, expected time.Time) error {
	if !actual.Before(expected) {
		return fmt.Errorf("time %v must be before %v", actual, expected)
	}
	return nil
}

// AssertWithinDuration checks if a time is within a duration of another time
func AssertWithinDuration(actual, expected time.Time, duration time.Duration) error {
	diff := actual.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	if diff > duration {
		return fmt.Errorf("time %v is not within %v of %v", actual, duration, expected)
	}
	return nil
}

// String Assertions

// AssertMatches checks if a string matches a regex pattern
func AssertMatches(value, pattern string) error {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %v", err)
	}
	if !matched {
		return fmt.Errorf("string %s does not match pattern %s", value, pattern)
	}
	return nil
}

// AssertStartsWith checks if a string starts with a prefix
func AssertStartsWith(value, prefix string) error {
	if !strings.HasPrefix(value, prefix) {
		return fmt.Errorf("string %s does not start with %s", value, prefix)
	}
	return nil
}

// AssertEndsWith checks if a string ends with a suffix
func AssertEndsWith(value, suffix string) error {
	if !strings.HasSuffix(value, suffix) {
		return fmt.Errorf("string %s does not end with %s", value, suffix)
	}
	return nil
}

// AssertContainsString checks if a string contains a substring
func AssertContainsString(value, substring string) error {
	if !strings.Contains(value, substring) {
		return fmt.Errorf("string %s does not contain %s", value, substring)
	}
	return nil
}
