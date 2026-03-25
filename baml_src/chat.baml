// baml_src/chat.baml
// 主对话 function

class Message {
  role string @description("user, assistant, or system")
  content string
}

type ToolsList = BashTool | ReadTool | WriteTool | EditTool | WebSearchTool | WebFetchTool

function Chat(messages: Message[], system_prompt: string) -> ChatResponse | ToolsList | UseSkill {
  client GLM5
  prompt #"
    {{ _.role("system") }}
    {{ system_prompt }}

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
    system_prompt: "You are a helpful assistant."
  }
}

test ChatChineseInput {
  functions [Chat]
  args {
    messages [
      { role: "user", content: "你好，请介绍一下你自己。" }
    ]
    system_prompt: "你是一个有帮助的助手。"
  }
  @client(GLM5)
}
