package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/redis/go-redis/v9"
)

type RateLimitInfo struct {
	Remaining int64 `json:"remaining"`
	Reset     int64 `json:"reset"`
}

func TokenBucketRateLimit(rdb *redis.Client, capacity, refillRate int64) graphql.HandlerExtension {
	return &tokenBucket{rdb, capacity, refillRate}
}

type tokenBucket struct {
	rdb        *redis.Client
	capacity   int64
	refillRate int64 // tokens per second
}

func (t *tokenBucket) ExtensionName() string { return "TokenBucketRateLimit" }

func (t *tokenBucket) Validate(graphql.ExecutableSchema) error { return nil }

func (t *tokenBucket) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	key := fmt.Sprintf("ratelimit:%s", t.clientKey(ctx))
	tokens, reset, err := t.consume(ctx, key)

	opCtx := graphql.GetOperationContext(ctx)
	opCtx.Headers.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(tokens, 0)))
	opCtx.Headers.Set("X-RateLimit-Reset", fmt.Sprintf("%d", reset))

	if err != nil {
		return func(ctx context.Context) *graphql.Response {
			println("Rate limiter error:", err)
			return graphql.ErrorResponse(ctx, "Rate limiter error")
		}
	}

	if tokens < 0 {
		return func(ctx context.Context) *graphql.Response {
			return graphql.ErrorResponse(ctx, "Rate limited. Try again later.")
		}
	}

	return next(ctx)
}

func (t *tokenBucket) clientKey(ctx context.Context) string {
	opCtx := graphql.GetOperationContext(ctx)
	xff := opCtx.Headers.Get("X-Forwarded-For")
	if xff != "" {
		first := strings.TrimSpace(strings.Split(xff, ",")[0])
		if first != "" {
			return first
		}
	}

	xri := strings.TrimSpace(opCtx.Headers.Get("X-Real-IP"))
	if xri != "" {
		return xri
	}

	return "global"
}

func (t *tokenBucket) consume(ctx context.Context, key string) (tokens, reset int64, err error) {
	now := time.Now().Unix()
	script := `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local tokens = redis.call('GET', key .. ':tokens')
if not tokens then tokens = capacity else tokens = tonumber(tokens) end

local last = redis.call('GET', key .. ':ts')
if last then
  local delta = math.min(now - tonumber(last), 3600)
  tokens = math.min(capacity, tokens + delta * rate)
end

if tokens >= 1 then
  tokens = tokens - 1
  redis.call('SET', key .. ':tokens', tokens)
  redis.call('SET', key .. ':ts', now)
  return tokens, now + math.ceil((1-tokens)/rate)
else
  return -1, 0
end`

	res, err := t.rdb.Eval(ctx, script, []string{key}, t.capacity, t.refillRate, now).Result()
	if err != nil {
		return 0, 0, err
	}

	values, ok := res.([]interface{})
	if !ok || len(values) < 2 {
		return 0, 0, errors.New("unexpected redis eval response")
	}

	tokens, err = toInt64(values[0])
	if err != nil {
		return 0, 0, err
	}

	reset, err = toInt64(values[1])
	if err != nil {
		return 0, 0, err
	}

	return tokens, reset, nil
}

func toInt64(v interface{}) (int64, error) {
	switch n := v.(type) {
	case int64:
		return n, nil
	case int:
		return int64(n), nil
	case float64:
		return int64(n), nil
	case string:
		return strconv.ParseInt(n, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported numeric type %T", v)
	}
}
