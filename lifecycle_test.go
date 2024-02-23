package lifecycle

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

func Test(t *testing.T) {
	lc := New()

	lc.Append(
		Hook{
			OnStart: func(ctx context.Context) error {
				fmt.Println("1: start")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				fmt.Println("1: stop")
				return nil
			},
		},
	)

	lc.Append(
		Hook{
			OnStart: func(ctx context.Context) error {
				return errors.New("2: start error")
			},
		},
	)

	lc.Append(
		Hook{
			OnStop: func(ctx context.Context) error {
				return errors.New("3: stop error")
			},
		},
	)

	if err := lc.Start(context.Background()); err != nil {
		fmt.Println(err)
		return
	}
	if err := lc.Stop(context.Background()); err != nil {
		fmt.Println(err)
		return
	}
}
