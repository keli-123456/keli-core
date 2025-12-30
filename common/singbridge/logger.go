package singbridge

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/sagernet/sing/common/logger"
	xerrors "github.com/xtls/xray-core/common/errors"
)

var _ logger.ContextLogger = (*XrayLogger)(nil)

type XrayLogger struct {
	newError func(values ...any) *xerrors.Error
}

func NewLogger(newErrorFunc func(values ...any) *xerrors.Error) *XrayLogger {
	return &XrayLogger{
		newErrorFunc,
	}
}

func (l *XrayLogger) Trace(args ...any) {
}

func (l *XrayLogger) Debug(args ...any) {
	xerrors.LogDebug(context.Background(), args...)
}

func (l *XrayLogger) Info(args ...any) {
	xerrors.LogInfo(context.Background(), args...)
}

func (l *XrayLogger) Warn(args ...any) {
	xerrors.LogWarning(context.Background(), args...)
}

func (l *XrayLogger) Error(args ...any) {
	if shouldDowngradeErrorArgs(args...) {
		xerrors.LogDebug(context.Background(), args...)
		return
	}
	xerrors.LogError(context.Background(), args...)
}

func (l *XrayLogger) Fatal(args ...any) {
}

func (l *XrayLogger) Panic(args ...any) {
}

func (l *XrayLogger) TraceContext(ctx context.Context, args ...any) {
}

func (l *XrayLogger) DebugContext(ctx context.Context, args ...any) {
	xerrors.LogDebug(ctx, args...)
}

func (l *XrayLogger) InfoContext(ctx context.Context, args ...any) {
	xerrors.LogInfo(ctx, args...)
}

func (l *XrayLogger) WarnContext(ctx context.Context, args ...any) {
	xerrors.LogWarning(ctx, args...)
}

func (l *XrayLogger) ErrorContext(ctx context.Context, args ...any) {
	if shouldDowngradeErrorArgs(args...) {
		xerrors.LogDebug(ctx, args...)
		return
	}
	xerrors.LogError(ctx, args...)
}

func (l *XrayLogger) FatalContext(ctx context.Context, args ...any) {
}

func (l *XrayLogger) PanicContext(ctx context.Context, args ...any) {
}

func shouldDowngradeErrorArgs(args ...any) bool {
	for _, arg := range args {
		if err, ok := arg.(error); ok {
			if shouldDowngradeError(err) {
				return true
			}
			continue
		}
		if s, ok := arg.(string); ok && isNoisyUDPEOF(s) {
			return true
		}
	}
	return false
}

func shouldDowngradeError(err error) bool {
	if err == nil {
		return false
	}
	if isNoisyUDPEOF(err.Error()) {
		return true
	}
	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)
}

func isNoisyUDPEOF(s string) bool {
	s = strings.ToLower(s)
	return strings.Contains(s, "decode udp message") && strings.Contains(s, "eof")
}
