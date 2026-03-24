// baml_src/chat.baml
// 主对话 function

class Message {
  role string @description("user, assistant, or system")
  content string
}

function Chat(messages: Message[], system_prompt: string) -> string {
  client GLM47
  prompt #"
    {{ _.role("system") }}
    {{ system_prompt }}

    {{ ctx.output_format }}

    {% for msg in messages %}
    {{ _.role(msg.role) }}
    {{ msg.content }}
    {% endfor %}
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
