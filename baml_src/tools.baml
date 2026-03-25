// baml_src/tools.baml
// Tool definitions for LLM tool selection

class ChatResponse {
  content string @description("The text response to the user")
}


class UseSkill {
  skillName string @description("Name of the skill to use")
  input string @description("Input for the skill")
}

//////////////////////
// Tool definitions

class BashTool {
  command string @description("The bash command to execute")
}

class EditTool {
  filePath string @description("Path to the file to edit")
  oldString string @description("String to replace")
  newString string @description("New string")
}

class ReadTool {
  filePath string @description("Path to the file to read")
}

class WebFetchTool {
  url string @description("URL to fetch")
}

class WebSearchTool {
  query string @description("Search query")
}

class WriteTool {
  filePath string @description("Path to the file to write")
  content string @description("Content to write")
}