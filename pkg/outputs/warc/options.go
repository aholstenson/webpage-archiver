package warc

type warcConfig struct {
	prefix string
}

type Option func(c *warcConfig)

func WithPrefix(prefix string) Option {
	return func(c *warcConfig) {
		c.prefix = prefix
	}
}
