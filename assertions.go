package assertions

import (
	"errors"
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
