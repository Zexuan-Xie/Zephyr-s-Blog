#!/usr/bin/env node
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { BlogBackendClient } from "./backendClient.js";
import { loadConfig } from "./config.js";
import { buildToolDefinitions } from "./tools.js";

export function createServer(): McpServer {
  const config = loadConfig();
  const client = new BlogBackendClient({ baseUrl: config.apiBaseUrl, adminToken: config.adminToken });
  const server = new McpServer({ name: "xlab-blog-mcp", version: "0.1.0" });

  for (const tool of buildToolDefinitions(config, client)) {
    server.registerTool(
      tool.name,
      {
        title: tool.title,
        description: tool.description,
        inputSchema: tool.inputSchema,
      },
      async (args) => tool.handler(args),
    );
  }

  return server;
}

export async function main(): Promise<void> {
  const server = createServer();
  const transport = new StdioServerTransport();
  await server.connect(transport);
}

if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch((error) => {
    console.error(error);
    process.exit(1);
  });
}
