package cron

import (
	rcron "github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("cron",
		fx.Provide(rcron.New),
		fx.Provide(NewSomeCron),
		fx.Invoke(
			func(lc fx.Lifecycle, c *Cron) {
				lc.Append(fx.Hook{
					OnStart: c.start,
					OnStop:  c.stop,
				})
			},
		),
	)
}
