package provider

// NewTogether creates a new Together AI provider
// Together uses the OpenAI-compatible API format
func NewTogether(cfg Config) *OpenAICompatProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.together.xyz/v1"
	}
	return NewOpenAICompat("together", cfg)
}
