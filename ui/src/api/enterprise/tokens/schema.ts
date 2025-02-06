import { z } from "zod";

const permissionTopics = [
  "namespaces",
  "secrets",
  "variables",
  "instances",
] as const; // TODO: finalize this array

const permissionMethods = [
  "POST",
  "GET",
  "DELETE",
  "PATCH",
  "PUT",
  "read",
  "manage",
] as const;

/**
 * the ui only offers a subset of the methods
 */
export const permissionMethodsAvailableUi: (typeof permissionMethods)[number][] =
  ["manage", "read"] as const;

const PermisionSchema = z.object({
  topic: z.enum(permissionTopics),
  method: z.enum(permissionMethods),
});

/**
 * example:
 * 
{
  "name": "foo1",
  "description": "foo1 description",
  "prefix": "832e0b8e",
  "permissions": [
    {
      "topic": "foo1_topic1",
      "method": "foo1_method1"
    },
    {
      "topic": "foo1_topic2",
      "method": "foo1_method2"
    }
  ],
  "createdAt": "2025-02-06T09:35:50.800122Z",
  "updatedAt": "2025-02-06T09:35:50.800122Z"
}
 */
const TokenSchema = z.object({
  name: z.string(),
  description: z.string(),
  prefix: z.string(),
  permissions: z.array(PermisionSchema),
  createdAt: z.string(),
  updatedAt: z.string(),
});

/**
 * example:
 * 
  {
    "data": [...]
  }
 */
export const TokenListSchema = z.object({
  data: z.array(TokenSchema),
});

/**
 * example:
 * 
  {
    "token": {...},
    "secret": "6dcbe0b0-f824-423c-be17-f199e57e1653"
  }
 */
export const TokenCreatedSchema = z.object({
  data: z.object({
    token: TokenSchema,
    secret: z.string(),
  }),
});

// TODO: is this still duration still needed?
export const ISO8601durationSchema = z
  .string()
  .regex(
    /^P(?!$)(\d+(?:\.\d+)?Y)?(\d+(?:\.\d+)?M)?(\d+(?:\.\d+)?W)?(\d+(?:\.\d+)?D)?(T(?=\d)(\d+(?:\.\d+)?H)?(\d+(?:\.\d+)?M)?(\d+(?:\.\d+)?S)?)?$/,
    {
      message: "Invalid ISO 8601 duration format",
    }
  );

/**
 * example
 * 
  {
    "name": "token name",
    "description": "",
    "permissions": [
      { "topic": "namespace", "method": "read" },
      { "topic": "files", "method": "manage" }
    ]
  }
 */
export const TokenFormSchema = z.object({
  name: z.string().nonempty(),
  description: z.string().nonempty(),
  duration: ISO8601durationSchema,
  permissions: z.array(PermisionSchema),
});

export const TokenDeletedSchema = z.null();

export type TokenSchemaType = z.infer<typeof TokenSchema>;
export type TokenListSchemaType = z.infer<typeof TokenListSchema>;
export type TokenFormSchemaType = z.infer<typeof TokenFormSchema>;
