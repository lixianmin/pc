// baml_src/chat.baml
// 主对话 function

class Message {
  role string @description("user, assistant, or system")
  content string
}

type ToolsList = Bash | Edit| Read | WebFetch | WebSearch | Write 

function Chat(systemPrompt: string, messages: Message[]) -> ChatResponse | ToolsList | UseSkill {
  client GLM5
  prompt #"
    {{ _.role("system") }}
    {{ systemPrompt }}

    {% for msg in messages %}
      {{ _.role(msg.role) }}
      {{ msg.content }}
    {% endfor %}

    {{ ctx.output_format }}
  "#
}

test ChatMultiTurn {
  functions [Chat]
  args {
    messages [
      { role: "user", content: "My name is Alice." },
      { role: "assistant", content: "Nice to meet you, Alice!" },
      { role: "user", content: "What is my name?" }
    ]
    systemPrompt: "You are a helpful assistant."
  }
}

test ChatChineseInput {
  functions [Chat]
  args {
    systemPrompt: "你是一个有帮助的助手。"

    messages [
      { role: "user", content: "你好，请介绍一下你自己。" }
    ]
  }
  @client(GLM5)
}
