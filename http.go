package stretcher

type HTTPOptions struct {
	Headers  map[string]string `yaml:"headers"`
	RetryMax int               `yaml:"retry_max"`
}
