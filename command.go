package healthplanet

import (
	"context"
	"io"
)

type runnerFn func(context.Context, []string, io.Writer, io.Writer) error

func (r runnerFn) run(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	return r(ctx, argv, outStream, errStream)
}

func runnerFunc(fn func(context.Context, []string, io.Writer, io.Writer) error) runner {
	return runnerFn(fn)
}
