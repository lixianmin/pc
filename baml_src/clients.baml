// baml_src/clients.baml
// LLM Client 配置

client<llm> Gpt4o {
  provider openai
  options {
    model "gpt-4o-mini"
    api_key env.OPENAI_API_KEY
  }
}

client<llm> Claude {
  provider anthropic
  options {
    model "claude-3-5-sonnet-20241022"
    api_key env.ANTHROPIC_API_KEY
  }
}

client<llm> Ollama {
  provider ollama
  options {
    model "llama3"
    base_url env.OLLAMA_BASE_URL
  }
}
