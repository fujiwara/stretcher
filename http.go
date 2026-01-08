package stretcher

type HTTPOptions struct {
	Headers map[string]string `help:"HTTP request headers(key=value) for download src archives with HTTP or HTTPS"`
}
