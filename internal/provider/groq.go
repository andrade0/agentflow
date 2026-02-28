package provider

// NewGroq creates a new Groq provider
// Groq uses the OpenAI-compatible API format
func NewGroq(cfg Config) *OpenAICompatProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.groq.com/openai/v1"
	}
	return NewOpenAICompat("groq", cfg)
}
