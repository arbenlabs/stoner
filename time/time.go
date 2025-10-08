package time

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TimeZone represents a timezone utility
type TimeZone struct {
	Location *time.Location
}

// NewTimeZone creates a new timezone utility
func NewTimeZone(timezone string) (*TimeZone, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone %s: %w", timezone, err)
	}

	return &TimeZone{Location: loc}, nil
}

// Now returns the current time in the timezone
func (tz *TimeZone) Now() time.Time {
	return time.Now().In(tz.Location)
}

// Parse parses a time string in the timezone
func (tz *TimeZone) Parse(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, tz.Location)
}

// Format formats a time in the timezone
func (tz *TimeZone) Format(t time.Time, layout string) string {
	return t.In(tz.Location).Format(layout)
}

// Date represents a date utility
type Date struct {
	Year  int
	Month int
	Day   int
}

// NewDate creates a new date
func NewDate(year, month, day int) Date {
	return Date{Year: year, Month: month, Day: day}
}

// FromTime creates a date from a time
func FromTime(t time.Time) Date {
	return Date{
		Year:  t.Year(),
		Month: int(t.Month()),
		Day:   t.Day(),
	}
}

// ToTime converts a date to time
func (d Date) ToTime() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
}

// String returns the string representation of the date
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// IsValid checks if the date is valid
func (d Date) IsValid() bool {
	t := d.ToTime()
	return t.Year() == d.Year && int(t.Month()) == d.Month && t.Day() == d.Day
}

// Duration represents a duration utility
type Duration struct {
	Duration time.Duration
}

// NewDuration creates a new duration
func NewDuration(d time.Duration) Duration {
	return Duration{Duration: d}
}

// FromString creates a duration from a string
func FromString(s string) (Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return Duration{}, fmt.Errorf("failed to parse duration: %w", err)
	}
	return Duration{Duration: d}, nil
}

// Add adds another duration
func (d Duration) Add(other Duration) Duration {
	return Duration{Duration: d.Duration + other.Duration}
}

// Subtract subtracts another duration
func (d Duration) Subtract(other Duration) Duration {
	return Duration{Duration: d.Duration - other.Duration}
}

// Multiply multiplies the duration by a factor
func (d Duration) Multiply(factor float64) Duration {
	return Duration{Duration: time.Duration(float64(d.Duration) * factor)}
}

// String returns the string representation of the duration
func (d Duration) String() string {
	return d.Duration.String()
}

// Cron represents a cron-like scheduler
type Cron struct {
	Jobs []Job
}

// Job represents a scheduled job
type Job struct {
	ID       string
	Schedule string
	Function func()
	LastRun  time.Time
	NextRun  time.Time
}

// NewCron creates a new cron scheduler
func NewCron() *Cron {
	return &Cron{
		Jobs: make([]Job, 0),
	}
}

// AddJob adds a job to the cron scheduler
func (c *Cron) AddJob(id, schedule string, fn func()) error {
	job := Job{
		ID:       id,
		Schedule: schedule,
		Function: fn,
	}

	// Parse schedule and calculate next run
	nextRun, err := c.parseSchedule(schedule)
	if err != nil {
		return fmt.Errorf("failed to parse schedule: %w", err)
	}

	job.NextRun = nextRun
	c.Jobs = append(c.Jobs, job)

	return nil
}

// Start starts the cron scheduler
func (c *Cron) Start() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			for i := range c.Jobs {
				if c.Jobs[i].NextRun.Before(now) || c.Jobs[i].NextRun.Equal(now) {
					c.Jobs[i].Function()
					c.Jobs[i].LastRun = now

					// Calculate next run
					nextRun, err := c.parseSchedule(c.Jobs[i].Schedule)
					if err == nil {
						c.Jobs[i].NextRun = nextRun
					}
				}
			}
		}
	}()
}

// parseSchedule parses a cron-like schedule (simplified version)
func (c *Cron) parseSchedule(schedule string) (time.Time, error) {
	// This is a simplified version - in production you'd want a full cron parser
	parts := strings.Fields(schedule)
	if len(parts) != 5 {
		return time.Time{}, fmt.Errorf("invalid schedule format")
	}

	// For now, just support "every X minutes" format
	if parts[0] == "*" && parts[1] == "*" && parts[2] == "*" && parts[3] == "*" {
		if parts[4] == "*" {
			return time.Now().Add(1 * time.Minute), nil
		}

		// Parse minutes
		if minutes, err := strconv.Atoi(parts[4]); err == nil {
			return time.Now().Add(time.Duration(minutes) * time.Minute), nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported schedule format")
}

// TimeCalculator provides time calculation utilities
type TimeCalculator struct{}

// NewTimeCalculator creates a new time calculator
func NewTimeCalculator() *TimeCalculator {
	return &TimeCalculator{}
}

// AddDays adds days to a time
func (tc *TimeCalculator) AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// AddMonths adds months to a time
func (tc *TimeCalculator) AddMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// AddYears adds years to a time
func (tc *TimeCalculator) AddYears(t time.Time, years int) time.Time {
	return t.AddDate(years, 0, 0)
}

// StartOfDay returns the start of the day
func (tc *TimeCalculator) StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day
func (tc *TimeCalculator) EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek returns the start of the week (Monday)
func (tc *TimeCalculator) StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 0, but we want it to be 7
	}
	return tc.StartOfDay(t.AddDate(0, 0, -(weekday - 1)))
}

// EndOfWeek returns the end of the week (Sunday)
func (tc *TimeCalculator) EndOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 0, but we want it to be 7
	}
	return tc.EndOfDay(t.AddDate(0, 0, 7-weekday))
}

// StartOfMonth returns the start of the month
func (tc *TimeCalculator) StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the end of the month
func (tc *TimeCalculator) EndOfMonth(t time.Time) time.Time {
	return tc.StartOfMonth(t.AddDate(0, 1, 0)).Add(-time.Nanosecond)
}

// StartOfYear returns the start of the year
func (tc *TimeCalculator) StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear returns the end of the year
func (tc *TimeCalculator) EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

// DaysBetween calculates the number of days between two times
func (tc *TimeCalculator) DaysBetween(start, end time.Time) int {
	startDay := tc.StartOfDay(start)
	endDay := tc.StartOfDay(end)
	return int(endDay.Sub(startDay).Hours() / 24)
}

// HoursBetween calculates the number of hours between two times
func (tc *TimeCalculator) HoursBetween(start, end time.Time) int {
	return int(end.Sub(start).Hours())
}

// MinutesBetween calculates the number of minutes between two times
func (tc *TimeCalculator) MinutesBetween(start, end time.Time) int {
	return int(end.Sub(start).Minutes())
}

// IsWeekend checks if a time falls on a weekend
func (tc *TimeCalculator) IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsWeekday checks if a time falls on a weekday
func (tc *TimeCalculator) IsWeekday(t time.Time) bool {
	return !tc.IsWeekend(t)
}

// FormatTime formats a time with common layouts
type FormatTime struct{}

// NewFormatTime creates a new time formatter
func NewFormatTime() *FormatTime {
	return &FormatTime{}
}

// RFC3339 formats time in RFC3339 format
func (ft *FormatTime) RFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ISO8601 formats time in ISO8601 format
func (ft *FormatTime) ISO8601(t time.Time) string {
	return t.Format("2006-01-02T15:04:05Z07:00")
}

// DateOnly formats time as date only
func (ft *FormatTime) DateOnly(t time.Time) string {
	return t.Format("2006-01-02")
}

// TimeOnly formats time as time only
func (ft *FormatTime) TimeOnly(t time.Time) string {
	return t.Format("15:04:05")
}

// DateTime formats time as date and time
func (ft *FormatTime) DateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// HumanReadable formats time in a human-readable format
func (ft *FormatTime) HumanReadable(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else {
		return t.Format("Jan 2, 2006")
	}
}

// ParseTime parses time from various formats
type ParseTime struct{}

// NewParseTime creates a new time parser
func NewParseTime() *ParseTime {
	return &ParseTime{}
}

// FromString parses time from a string using common formats
func (pt *ParseTime) FromString(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"15:04:05",
		"Jan 2, 2006",
		"January 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

// FromUnix parses time from Unix timestamp
func (pt *ParseTime) FromUnix(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// FromUnixNano parses time from Unix nanosecond timestamp
func (pt *ParseTime) FromUnixNano(timestamp int64) time.Time {
	return time.Unix(0, timestamp)
}
