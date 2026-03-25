// baml_src/tools.baml
// Tool definitions for LLM tool selection
// https://zhuanlan.zhihu.com/p/1917869825310557841

class ChatResponse {
  apiName "chat_response"
  content string @description("The final text response to user")
}


class UseSkill {
  apiName "use_skill"
  skillName string @description("Name of the skill to use")
  input string @description("Input for the skill")
}

//////////////////////
// Tool definitions

class Bash {
  apiName "bash"
  command string @description("The command to execute")
  description string? @description("Clear, concise description of what this command does in 5-10 words. Examples:\nInput: ls\nOutput: Lists files in current directory\n\nInput: git status\nOutput: Shows working tree status\n\nInput: npm install\nOutput: Installs package dependencies\n\nInput: mkdir foo\nOutput: Creates directory 'foo'")
  timeout int? @description("Optional timeoutin milliseconds (max 600000)")
}

class Edit {
  apiName "edit_file"
  filePath string @description("The absolute path to the file to modify")
  oldString string @description("The text to replace")
  newString string @description("The text to replace it with (must be different from old_string)")
  expectedReplacements int? @description("The expected number of replacements to perform. Defaults to 1 if not specified.")
}

class Read {
  apiName "read_file"
  filePath string @description("The absolute path to the file to read")
  limit int? @description("The number of lines to read. Only provide if the file is too large to read at once")
  offset int? @description("The line number to start reading from. Only provide if the file is too large to read at once")
}

class WebFetch {
  apiName "web_fetch"
  url string @description("The URL to fetch content from")
  prompt string @description("The prompt to run on the fetched content")
}

class WebSearch {
  apiName "web_search"
  query string @description("The search query to use")
  allowedDomains string[]? @description("Only include search results from these domains")
  blocked_domains string[]? @description("Never include search results from these domains")
}

class Write {
  apiName "write_file"
  filePath string @description("The absolute path to the file to write (must be absolute, not relative)")
  content string @description("The content to write to the file")
}