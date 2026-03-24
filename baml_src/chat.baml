// baml_src/chat.baml
// 主对话 function

class Message {
  role string @description("user, assistant, or system")
  content string
}

function Chat(messages: Message[], system_prompt: string) -> string {
  client "Gpt4o"
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
