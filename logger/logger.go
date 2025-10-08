package logger

import (
	"errors"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"context"

	"os"
)

// **************************************************
// Logger
// Logger is a wrapper around the slog.Logger default implementation.
// It is used to configure and create a logger with a specific config.
// **************************************************

type Logger struct {
	*slog.Logger
	config *LoggerConfig
}

type LoggerConfig struct {
	Level              slog.Level
	AddSource          bool
	ServiceName        string
	ServiceVersion     string
	ServiceEnvironment string
	Writer             io.Writer
}

var defaultLogger *Logger

// NewLogger creates a new logger with a specific logger config.
func NewLogger(config *LoggerConfig) (*Logger, error) {
	if config == nil {
		return nil, errors.New("logger config is required")
	}

	opts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {

			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)

				parts := strings.Split(source.File, "/")
				source.File = parts[len(parts)-1]
			}

			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   "timestamp",
					Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
				}
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(config.Writer, opts)
	logger := slog.New(handler).With(
		slog.String("service.name", config.ServiceName),
		slog.String("service.version", config.ServiceVersion),
		slog.String("service.environment", config.ServiceEnvironment),
	).WithGroup("service")

	slog.SetDefault(logger)
	logs := &Logger{
		Logger: logger,
		config: config,
	}

	defaultLogger = logs

	return logs, nil
}

// NewLoggerConfig creates a new logger config with a specific service name, add source, service version, and service environment.
func NewLoggerConfig(serviceName string, shouldAddSource bool, serviceVersion string, serviceEnvironment string) *LoggerConfig {
	return &LoggerConfig{
		Level:              slog.LevelInfo,
		AddSource:          shouldAddSource,
		ServiceName:        serviceName,
		ServiceVersion:     serviceVersion,
		ServiceEnvironment: serviceEnvironment,
		Writer:             os.Stdout,
	}
}

type contextKey string

const (
	TraceIDKey   contextKey = "trace_id"
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
	SessionIDKey contextKey = "session_id"
)

// withSlog wraps the slog.Logger with additional functionality.
func (l *Logger) withSlog(args ...interface{}) *slog.Logger {
	return l.Logger.With(args...)
}

// WithTraceID wraps the slog.Logger with a trace ID.
func (l *Logger) WithTraceID(traceID string) *Logger {
	return &Logger{
		Logger: l.withSlog("trace_id", traceID),
		config: l.config,
	}
}

// WithRequestID wraps the slog.Logger with a request ID.
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.withSlog("request_id", requestID),
		config: l.config,
	}
}

// WithUserID wraps the slog.Logger with a user ID.
func (l *Logger) WithUserID(userID interface{}) *Logger {
	return &Logger{
		Logger: l.withSlog("user_id", userID),
		config: l.config,
	}
}

// WithFields wraps the slog.Logger with a map of fields.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		Logger: l.withSlog(args...),
		config: l.config,
	}
}

// WithComponent wraps the slog.Logger with a component.
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.withSlog("component", component),
		config: l.config,
	}
}

// WithContext wraps the slog.Logger with a context.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.Logger

	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		logger = logger.With("trace_id", traceID)
	}

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		logger = logger.With("request_id", requestID)
	}

	if userID := ctx.Value(UserIDKey); userID != nil {
		logger = logger.With("user_id", userID)
	}

	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok && sessionID != "" {
		logger = logger.With("session_id", sessionID)
	}

	return &Logger{
		Logger: logger,
		config: l.config,
	}
}

// InfoIf logs an info message if the condition is true.
func (l *Logger) InfoIf(condition bool, msg string, fields ...interface{}) {
	if condition {
		l.Info(msg, fields...)
	}
}

// WarnIf logs a warn message if the condition is true.
func (l *Logger) WarnIf(condition bool, msg string, fields ...interface{}) {
	if condition {
		l.Warn(msg, fields...)
	}
}

// ErrorIf logs an error message if the condition is true.
func (l *Logger) ErrorIf(condition bool, msg string, fields ...interface{}) {
	if condition {
		l.Error(msg, fields...)
	}
}

// ErrorWithStack logs an error message with a stack trace.
func (l *Logger) ErrorWithStack(msg string, err error, fields ...interface{}) {
	args := []interface{}{"error", err.Error()}
	args = append(args, fields...)

	if l.config.Level <= slog.LevelDebug {
		stack := make([]byte, 4096)
		length := runtime.Stack(stack, false)
		args = append(args, "stack_trace", string(stack[:length]))
	}

	l.Error(msg, args...)
}

// LogError logs an error message with a structured error details.
type ErrorDetails struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (l *Logger) LogError(err error, details ErrorDetails, fields ...interface{}) {
	args := []interface{}{
		"error", err.Error(),
		"error_code", details.Code,
		"error_message", details.Message,
	}

	if details.Details != nil {
		args = append(args, "error_details", details.Details)
	}

	args = append(args, fields...)
	l.Error("Structured error", args...)
}

// GetLogger gets the default logger.
func GetLogger() *Logger {
	if defaultLogger == nil {
		panic("logger is not initialized, no default logger found")
	}
	return defaultLogger
}

// Debug logs a debug message.
func Debug(msg string, fields ...interface{}) {
	GetLogger().Debug(msg, fields...)
}

// Info logs an info message.
func Info(msg string, fields ...interface{}) {
	GetLogger().Info(msg, fields...)
}

// Warn logs a warn message.
func Warn(msg string, fields ...interface{}) {
	GetLogger().Warn(msg, fields...)
}

// Error logs an error message.
func Error(msg string, fields ...interface{}) {
	GetLogger().Error(msg, fields...)
}

// WithTraceID wraps the slog.Logger with a trace ID.
func WithTraceID(traceID string) *Logger {
	return GetLogger().WithTraceID(traceID)
}

// WithRequestID wraps the slog.Logger with a request ID.
func WithRequestID(requestID string) *Logger {
	return GetLogger().WithRequestID(requestID)
}

// WithContext wraps the slog.Logger with a context.
func WithContext(ctx context.Context) *Logger {
	return GetLogger().WithContext(ctx)
}

// WithFields wraps the slog.Logger with a map of fields.
func WithFields(fields map[string]interface{}) *Logger {
	return GetLogger().WithFields(fields)
}

// ErrorWithStack logs an error message with a stack trace.
func ErrorWithStack(msg string, err error, fields ...interface{}) {
	GetLogger().ErrorWithStack(msg, err, fields...)
}

// ContextWithTraceID wraps the context.Context with a trace ID.
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// ContextWithRequestID wraps the context.Context with a request ID.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// ContextWithUserID wraps the context.Context with a user ID.
func ContextWithUserID(ctx context.Context, userID interface{}) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
