package gong

type componentConfig struct {
	swap    string
	trigger string
}

type ComponentOption func(config componentConfig) componentConfig

func WithSwap(swap string) ComponentOption {
	return func(config componentConfig) componentConfig {
		config.swap = swap
		return config
	}
}

func WithTrigger(trigger string) ComponentOption {
	return func(config componentConfig) componentConfig {
		config.trigger = trigger
		return config
	}
}
