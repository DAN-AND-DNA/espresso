package ctxhelper

import (
	"context"
	"github.com/dan-and-dna/minilog"
	"github.com/gin-gonic/gin"
	"github.com/kamilsk/tracer"
	"github.com/prometheus/client_golang/prometheus"
)

func Set(ctx context.Context, key string, value any) context.Context {
	if c, ok := ctx.(*gin.Context); ok {
		c.Set(key, value)
	} else {
		ctx = context.WithValue(ctx, key, value)
	}

	return ctx
}

func Get(ctx context.Context, key string) any {
	if c, ok := ctx.(*gin.Context); ok {
		val, exists := c.Get(key)
		if exists {
			return val
		}
		return nil
	}
	return ctx.Value(key)
}

const TraceKey = "core_trace_key"

func InjectTrace(ctx context.Context, trace *tracer.Trace) context.Context {
	return Set(ctx, TraceKey, trace)
}

func FetchTrace(ctx context.Context) *tracer.Trace {
	ptr := Get(ctx, TraceKey)
	if ptr == nil {
		return nil
	}

	ret, _ := ptr.(*tracer.Trace)
	return ret
}

const LoggerKey = "core_logger_key"

func InjectLogger(ctx context.Context, logger *minilog.MiniLog) context.Context {
	return Set(ctx, LoggerKey, logger)
}

func FetchLogger(ctx context.Context) *minilog.MiniLog {
	ptr := Get(ctx, LoggerKey)
	if ptr == nil {
		return nil
	}

	ret, _ := ptr.(*minilog.MiniLog)
	return ret
}

const RegistryKey = "core_registry_key"

func InjectRegistry(ctx context.Context, registry *prometheus.Registry) context.Context {
	return Set(ctx, RegistryKey, registry)
}

func FetchRegistry(ctx context.Context) *prometheus.Registry {
	ptr := Get(ctx, RegistryKey)
	if ptr == nil {
		return nil
	}

	ret, _ := ptr.(*prometheus.Registry)
	return ret
}

const RequestIdKey = "core_requestId"

func InjectRequestId(ctx context.Context, requestId string) context.Context {
	return Set(ctx, RequestIdKey, requestId)
}

func FetchRequestId(ctx context.Context) string {
	ptr := Get(ctx, RequestIdKey)
	if ptr == nil {
		return ""
	}

	ret, _ := ptr.(string)
	return ret
}

func FetchClientIP(ctx context.Context) string {
	if c, ok := ctx.(*gin.Context); ok {
		return c.ClientIP()
	}
	return ""
}

const MetadataKey = "core_metadata"

func InjectMetadata(ctx context.Context, metadata map[string][]string) context.Context {
	return Set(ctx, MetadataKey, metadata)
}

func FetchMetadata(ctx context.Context) map[string][]string {
	val := Get(ctx, MetadataKey)
	if val != nil {
		return val.(map[string][]string)
	}

	return nil
}
