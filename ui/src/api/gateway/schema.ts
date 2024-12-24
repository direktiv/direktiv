import { z } from "zod";

export const routeMethods = [
  "CONNECT",
  "DELETE",
  "GET",
  "HEAD",
  "OPTIONS",
  "PATCH",
  "POST",
  "PUT",
  "TRACE",
] as const;

export const MethodsSchema = z.enum(routeMethods);

/**
 * example
  {
    "type": "proxy",
    "configuration": {
      "key1": ["value1", "value2", "value3"],
      "key2": "value3"
    }
  }
 */
const PluginSchema = z.object({
  type: z.string(),
  configuration: z.record(z.unknown()).optional(),
});

/**
 * example
{
  "file_path": "/test/roottree.yaml",
  "path": "/adon/",
  "server_path": "/ns/git-test/adon",
  "methods": [
    "GET"
  ],
  "allow_anonymous": true,
  "timeout": 0,
  "errors": [],
  "warnings": [],
  "plugins": {
    "target": {
      "type": "instant-response",
      "configuration": {
        "content_type": "application/xml",
        "status_code": 200,
        "status_message": "..."
      }
    }
  }
}
 */

const filterInvalidEntries = (schema: z.ZodTypeAny) =>
  z
    .array(z.any())
    .transform((entryArr) =>
      entryArr.filter((entry) => schema.safeParse(entry).success)
    );

const RouteSchema = z.object({
  methods: filterInvalidEntries(MethodsSchema).nullable(),
  file_path: z.string(),
  path: z.string().optional(),
  server_path: z.string().optional(),
  allow_anonymous: z.boolean(),
  errors: z.array(z.string()),
  warnings: z.array(z.string()),
  plugins: z.object({
    // if a user might use an unsupported plugin, they will be parsed out instead of throwing an error
    outbound: filterInvalidEntries(PluginSchema).optional(),
    inbound: filterInvalidEntries(PluginSchema).optional(),
    auth: filterInvalidEntries(PluginSchema).optional(),
    target: PluginSchema.optional(),
  }),
});

export type RouteSchemaType = z.infer<typeof RouteSchema>;

/**
 * example
  {
    "data": [{...}, {...}, {...}]
  } 
 */
export const RoutesListSchema = z.object({
  data: z.array(RouteSchema),
});

export type RoutesListSchemaType = z.infer<typeof RoutesListSchema>;

/**
 * example
{
  "data": [
    {
      "username": "user",
      "password": "pwd",
      "api_key": "key",
      "tags": [
        "tag1"
      ],
      "groups": [
        "group1"
      ]
    }
  ]
}
 */
const ConsumerSchema = z.object({
  username: z.string(),
  password: z.string(),
  api_key: z.string(),
  tags: z.array(z.string()).nullable(),
  groups: z.array(z.string()).nullable(),
});

export type ConsumerSchemaType = z.infer<typeof ConsumerSchema>;

/**
 * example
  {
    "data": [{...}, {...}, {...}]
  } 
 */
export const ConsumersListSchema = z.object({
  data: z.array(ConsumerSchema),
});

// OpenAPI Info Schema
const OpenAPIInfoSchema = z.object({
  title: z.string(),
  description: z.string(),
  version: z.string(),
});

// OpenAPI Server Schema
const OpenAPIServerSchema = z.object({
  url: z.string(),
  description: z.string(),
});

// OpenAPI Response Schema
const ResponseSchema = z.object({
  description: z.string(),
});

// OpenAPI Method Schema (GET, POST, etc.)
const OpenAPIMethodSchema = z.object({
  summary: z.string().optional(),
  description: z.string().optional(),
  responses: z.record(ResponseSchema), // A record of HTTP status codes and responses
});

// OpenAPI Path Item Schema
const PathItemSchema = z.record(
  z.string(),
  z.union([
    OpenAPIMethodSchema, // GET, POST, etc.
    z.object({}), // For cases where there are no methods defined
  ])
);

// OpenAPI Paths Schema
const PathsSchema = z.record(z.string(), PathItemSchema);

// Full OpenAPI Spec Schema
export const DocsSchema = z.object({
  data: z.object({
    openapi: z.string(),
    info: OpenAPIInfoSchema,
    servers: z.array(OpenAPIServerSchema),
    paths: PathsSchema,
  }),
});

export type DocsSchemaType = z.infer<typeof DocsSchema>;
