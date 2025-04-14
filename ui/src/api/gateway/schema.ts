import { z } from "zod";

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

const filterInvalidEntries = (schema: z.ZodTypeAny) =>
  z
    .array(z.any())
    .transform((entryArr) =>
      entryArr.filter((entry) => schema.safeParse(entry).success)
    );

/**
 * example
  {
    "inbound": [
      {
        "configuration": {
          "script": "some script"
        },
        "type": "js-inbound"
      },
      {
        "configuration": {
          "omit_body": false,
          "omit_consumer": false,
          "omit_headers": true,
          "omit_queries": true
        },
        "type": "request-convert"
      }
    ],
    "outbound": [],
    "target": {
      "configuration": {
        "status_code": 200
      },
      "type": "instant-response"
    }
  }
 */
const PluginsSchema = z.object({
  inbound: filterInvalidEntries(PluginSchema).optional(),
  outbound: filterInvalidEntries(PluginSchema).optional(),
  auth: filterInvalidEntries(PluginSchema).optional(),
  target: PluginSchema.optional(),
});

export type PluginsSchemaType = z.infer<typeof PluginsSchema>;

export type PluginType = keyof PluginsSchemaType;

export const OperationSchema = z.record(z.any());

export const MethodsSchema = z.object({
  connect: OperationSchema.optional(),
  delete: OperationSchema.optional(),
  get: OperationSchema.optional(),
  head: OperationSchema.optional(),
  options: OperationSchema.optional(),
  patch: OperationSchema.optional(),
  post: OperationSchema.optional(),
  put: OperationSchema.optional(),
  trace: OperationSchema.optional(),
});

export type MethodsSchemaType = z.infer<typeof MethodsSchema>;
export type RouteMethod = keyof MethodsSchemaType;

export const routeMethods: Set<RouteMethod> = new Set([
  "connect",
  "delete",
  "get",
  "head",
  "options",
  "patch",
  "post",
  "put",
  "trace",
]);

/**
 * example
{
  "x-direktiv-api": "endpoint/v2",
  "get": {...},
  "post": {...},
  "x-direktiv-config": {
    "allow_anonymous": true,
    "path": "path",
    "plugins": {...}
  }
}
 */
const DirektivOpenApiSpecSchema = z
  .object({
    "x-direktiv-api": z.literal("endpoint/v2").optional(),
    "x-direktiv-config": z
      .object({
        allow_anonymous: z.boolean().optional(),
        path: z.string().optional(),
        plugins: PluginsSchema.optional(),
      })
      .optional(),
  })
  .merge(MethodsSchema);

export type DirektivOpenApiSpecSchemaType = z.infer<
  typeof DirektivOpenApiSpecSchema
>;

export type MethodsKeys = keyof DirektivOpenApiSpecSchemaType;
export type MethodsObject = Partial<
  Pick<DirektivOpenApiSpecSchemaType, MethodsKeys>
>;

/**
 * example
  {
    "spec": {...},
    "file_path": "/route.yaml",
    "errors": [],
    "server_path": "/ns/demo/path",
    "warnings": []
  }
 */
export const RouteSchema = z.object({
  spec: DirektivOpenApiSpecSchema,
  file_path: z.string(),
  errors: z.array(z.string()),
  server_path: z.string(),
  warnings: z.array(z.string()),
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

/**
 * example INFO
{
  "data": {
    "spec": {
      "openapi": "3.0.0",
      "info": {
        "title": "cg",
        "version": "1.0"
      },
      "paths": {}
    },
    "file_path": "virtual",
    "errors": []
  }
}
 */
export const OpenapiSpecificationSchema = z.object({
  data: z.object({
    spec: z
      .object({
        openapi: z.string(),
        info: z
          .object({
            title: z.string(),
            version: z.string(),
            description: z.string().optional(),
          })
          .passthrough(),
        paths: z.record(z.any()),
      })
      .passthrough(),
    file_path: z.string(),
    errors: z.array(z.string()),
  }),
});

export type OpenapiSpecificationSchemaType = z.infer<
  typeof OpenapiSpecificationSchema
>;
