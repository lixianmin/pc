// baml_src/tools.baml
// Tool definitions for LLM tool selection

class BashTool {
  tool_name "bash" @description("Execute bash commands")
  command string @description("The bash command to execute")
}

class ReadTool {
  tool_name "read" @description("Read file content")
  file_path string @description("Path to the file to read")
}

class WriteTool {
  tool_name "write" @description("Write content to file")
  file_path string @description("Path to the file to write")
  content string @description("Content to write")
}

class EditTool {
  tool_name "edit" @description("Edit file by replacing old string with new string")
  file_path string @description("Path to the file to edit")
  old_string string @description("String to replace")
  new_string string @description("New string")
}

class WebSearchTool {
  tool_name "web_search" @description("Search the web for information")
  query string @description("Search query")
}

class WebFetchTool {
  tool_name "web_fetch" @description("Fetch content from a URL")
  url string @description("URL to fetch")
}

class UseSkill {
  tool_name "use_skill" @description("Use a skill by name with input")
  skill_name string @description("Name of the skill to use")
  input string @description("Input for the skill")
}

function ChooseTool(query: string, additional_context: string) -> BashTool | ReadTool | WriteTool | EditTool | WebSearchTool | WebFetchTool | UseSkill {
  client "openai/gpt-5"
  prompt #"
    Given the user query and context, determine which tool to use.

    {{ ctx.output_format }} 

    {{ _.role('user') }}
    Query: {{ query }}
    Context: {{ additional_context }}
  "#
}
