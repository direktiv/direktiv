import { z } from "zod";

/**
 * example:
 * 
  {
    "id": "13cbe5a1-3bc7-4f13-b5aa-658b046dabb4",
    "description": "my first token",
    "permissions": ["workflowView", "permissionsView"],
    "created": "2023-08-30T07:27:35.296195769Z",
    "expires": "2024-08-30T07:27:35.29614121Z",
    "expired": false
  }
 */
const TokenSchema = z.object({
  id: z.string(),
  description: z.string(),
  permissions: z.array(z.string()),
  created: z.string(),
  expires: z.string(),
  expired: z.boolean(),
});

/**
 * example:
 * 
  {
    "tokens": [...]
  }
 */
export const TokenListSchema = z.object({
  tokens: z.array(TokenSchema),
});

export const TokenCreatedSchema = z.null();

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
    "description": "my first token",
    "duration": "P1Y",
    "permissions": ["permissionsView", "workflowView"]
  }
 */
export const TokenFormSchema = z.object({
  description: z.string(),
  duration: ISO8601durationSchema,
  permissions: z.array(z.string()),
});

export type TokenSchemaType = z.infer<typeof TokenSchema>;
export type TokenFormSchemaType = z.infer<typeof TokenFormSchema>;
