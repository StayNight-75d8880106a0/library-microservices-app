package helper

import "context"

type traceCTXKey struct{}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceCTXKey{}, traceID)
}

func TraceIDFromContext(ctx context.Context) string {

	traceID, ok := ctx.Value(traceCTXKey{}).(string)

	if !ok {
		return ""
	}

	return traceID

}
