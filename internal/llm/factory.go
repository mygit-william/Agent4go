package llm

import (
	"fmt"
)

// Factory LLM 工厂
type Factory struct {
	config Config
}

// NewFactory 创建工厂
func NewFactory(config Config) *Factory {
	return &Factory{config: config}
}

// Make 创建 LLM 实例
func (f *Factory) Make(providerKey ...string) (Interface, error) {
	key := f.config.DefaultProvider
	if len(providerKey) > 0 && providerKey[0] != "" {
		key = providerKey[0]
	}

	provider, ok := f.config.Providers[key]
	if !ok {
		return nil, fmt.Errorf("provider '%s' 不存在", key)
	}

	switch provider.Driver {
	case "zhipu":
		return NewZhipuAdapter(provider.BaseURL, provider.Model, provider.APIKey)
	case "openai":
		return NewOpenAIAdapter(provider.BaseURL, provider.Model, provider.APIKey)
	case "ollama":
		return NewOllamaAdapter(provider.BaseURL, provider.Model)
	case "longcat":
		return NewLongcatAdapter(provider.BaseURL, provider.Model, provider.APIKey)
	default:
		return nil, fmt.Errorf("不支持的驱动: %s", provider.Driver)
	}
}

// GetAvailableProviders 获取可用提供商
func (f *Factory) GetAvailableProviders() []string {
	providers := make([]string, 0, len(f.config.Providers))
	for k := range f.config.Providers {
		providers = append(providers, k)
	}
	return providers
}

// GetDefaultProvider 获取默认提供商
func (f *Factory) GetDefaultProvider() string {
	return f.config.DefaultProvider
}
