package tokenv1

import (
	"github.com/braumsmilk/go-registry"
)

const RedisTokenValidityMinutes string = "token_validity_minutes"

func NewConfig() TokenConfig {
	regRedis := registry.Get().Redis

	validityHours := 12
	if regRedis.Extra != nil {
		validityMinutesVal, ok := regRedis.Extra[RedisTokenValidityMinutes]
		if ok {
			validityHours = validityMinutesVal.(int)
		}
	}

	return TokenConfig{
		ValidityHours: validityHours,
	}
}

type TokenConfig struct {
	ValidityHours int `yaml:"validity_hours"`
}
