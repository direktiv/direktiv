import { PermissionsArray } from "../schema";
import { z } from "zod";

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
  "isExpired": false,
  "expiredAt": "2025-02-06T09:35:50.800122Z",
  "createdAt": "2025-02-06T09:35:50.800122Z",
  "updatedAt": "2025-02-06T09:35:50.800122Z"
}
 */
const TokenSchema = z.object({
  name: z.string(),
  description: z.string(),
  prefix: z.string(),
  permissions: PermissionsArray,
  isExpired: z.boolean(),
  expiredAt: z.string(),
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
    "data": {
      "apiToken": {...},
      "secret": "6dcbe0b0-f824-423c-be17-f199e57e1653"
    }
  }
 */
export const TokenCreatedSchema = z.object({
  data: z.object({
    apiToken: TokenSchema,
    secret: z.string(),
  }),
});

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
    "description": "token description",
    "permissions": [
      { "topic": "namespace", "method": "read" },
      { "topic": "files", "method": "manage" }
    ]
  }
 */
export const TokenFormSchema = z.object({
  name: z.string().nonempty(),
  description: z.string(),
  duration: ISO8601durationSchema,
  permissions: PermissionsArray,
});

export const TokenDeletedSchema = z.null();

export type TokenSchemaType = z.infer<typeof TokenSchema>;
export type TokenListSchemaType = z.infer<typeof TokenListSchema>;
export type TokenFormSchemaType = z.infer<typeof TokenFormSchema>;
