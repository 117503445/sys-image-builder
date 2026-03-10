package main

import (
	"context"

	"github.com/rs/zerolog/log"
)

func build() {
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)
	log.Ctx(ctx).Info().Msg("building system image")
}
