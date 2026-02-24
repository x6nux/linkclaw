package llm

// 定价表（microdollars per token，即 USD × 1,000,000 / 1,000,000 = USD per token）
// 单位：microdollars per 1K tokens

type ModelPricing struct {
	InputPerMTok         int64 // 每百万 input token 的微美元数
	OutputPerMTok        int64
	CacheWritePerMTok    int64 // cache creation
	CacheReadPerMTok     int64 // cache read（最便宜）
	CachedInputPerMTok   int64 // OpenAI cached prompt
}

// anthropicPricing Anthropic 模型定价（截至 2025-02）
// https://www.anthropic.com/pricing
var anthropicPricing = map[string]ModelPricing{
	// claude-opus-4-x
	"claude-opus-4-5":              {InputPerMTok: 15_000_000, OutputPerMTok: 75_000_000, CacheWritePerMTok: 18_750_000, CacheReadPerMTok: 1_500_000},
	"claude-opus-4-0":              {InputPerMTok: 15_000_000, OutputPerMTok: 75_000_000, CacheWritePerMTok: 18_750_000, CacheReadPerMTok: 1_500_000},
	// claude-sonnet-4-x
	"claude-sonnet-4-6":            {InputPerMTok: 3_000_000, OutputPerMTok: 15_000_000, CacheWritePerMTok: 3_750_000, CacheReadPerMTok: 300_000},
	"claude-sonnet-4-5":            {InputPerMTok: 3_000_000, OutputPerMTok: 15_000_000, CacheWritePerMTok: 3_750_000, CacheReadPerMTok: 300_000},
	// claude-haiku-4-x
	"claude-haiku-4-5":             {InputPerMTok: 800_000, OutputPerMTok: 4_000_000, CacheWritePerMTok: 1_000_000, CacheReadPerMTok: 80_000},
	// claude-3-5 系列（老版本）
	"claude-3-5-sonnet-20241022":   {InputPerMTok: 3_000_000, OutputPerMTok: 15_000_000, CacheWritePerMTok: 3_750_000, CacheReadPerMTok: 300_000},
	"claude-3-5-haiku-20241022":    {InputPerMTok: 800_000, OutputPerMTok: 4_000_000, CacheWritePerMTok: 1_000_000, CacheReadPerMTok: 80_000},
	"claude-3-opus-20240229":       {InputPerMTok: 15_000_000, OutputPerMTok: 75_000_000, CacheWritePerMTok: 18_750_000, CacheReadPerMTok: 1_500_000},
}

// openaiPricing OpenAI 模型定价
var openaiPricing = map[string]ModelPricing{
	"gpt-4o":              {InputPerMTok: 2_500_000, OutputPerMTok: 10_000_000, CachedInputPerMTok: 1_250_000},
	"gpt-4o-mini":         {InputPerMTok: 150_000, OutputPerMTok: 600_000, CachedInputPerMTok: 75_000},
	"o1":                  {InputPerMTok: 15_000_000, OutputPerMTok: 60_000_000, CachedInputPerMTok: 7_500_000},
	"o1-mini":             {InputPerMTok: 1_100_000, OutputPerMTok: 4_400_000, CachedInputPerMTok: 550_000},
	"o3":                  {InputPerMTok: 10_000_000, OutputPerMTok: 40_000_000, CachedInputPerMTok: 2_500_000},
	"o3-mini":             {InputPerMTok: 1_100_000, OutputPerMTok: 4_400_000, CachedInputPerMTok: 275_000},
	"gpt-4.1":             {InputPerMTok: 2_000_000, OutputPerMTok: 8_000_000, CachedInputPerMTok: 500_000},
	"gpt-4.1-mini":        {InputPerMTok: 400_000, OutputPerMTok: 1_600_000, CachedInputPerMTok: 100_000},
}

// CalcAnthropicCost 计算 Anthropic 请求费用（返回微美元）
func CalcAnthropicCost(model string, usage AnthropicUsage) int64 {
	p, ok := anthropicPricing[model]
	if !ok {
		// 未知模型按 sonnet 价格估算
		p = anthropicPricing["claude-sonnet-4-6"]
	}
	cost := int64(usage.InputTokens)*p.InputPerMTok/1_000_000 +
		int64(usage.OutputTokens)*p.OutputPerMTok/1_000_000 +
		int64(usage.CacheCreationInputTokens)*p.CacheWritePerMTok/1_000_000 +
		int64(usage.CacheReadInputTokens)*p.CacheReadPerMTok/1_000_000
	return cost
}

// CalcOpenAICost 计算 OpenAI 请求费用（返回微美元）
func CalcOpenAICost(model string, usage OpenAIUsage) int64 {
	p, ok := openaiPricing[model]
	if !ok {
		p = openaiPricing["gpt-4o"]
	}
	cachedTokens := 0
	if usage.PromptTokensDetails != nil {
		cachedTokens = usage.PromptTokensDetails.CachedTokens
	}
	nonCachedInput := usage.PromptTokens - cachedTokens
	cost := int64(nonCachedInput)*p.InputPerMTok/1_000_000 +
		int64(cachedTokens)*p.CachedInputPerMTok/1_000_000 +
		int64(usage.CompletionTokens)*p.OutputPerMTok/1_000_000
	return cost
}

// MicrodollarsToUSD 微美元转 USD 字符串
func MicrodollarsToUSD(microdollars int64) float64 {
	return float64(microdollars) / 1_000_000.0
}

// KnownModels 返回已知模型列表（供前端下拉）
func KnownModels() map[ProviderType][]string {
	anthropicModels := make([]string, 0, len(anthropicPricing))
	for k := range anthropicPricing {
		anthropicModels = append(anthropicModels, k)
	}
	openaiModels := make([]string, 0, len(openaiPricing))
	for k := range openaiPricing {
		openaiModels = append(openaiModels, k)
	}
	return map[ProviderType][]string{
		ProviderAnthropic: anthropicModels,
		ProviderOpenAI:    openaiModels,
	}
}
