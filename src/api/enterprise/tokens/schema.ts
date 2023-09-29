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

/**
 * example:
 * 
  {
    "id": "7eff49c1-ec13-4d81-8278-9ad8e15ff1f5",
    "token": "6656c247e8cb6cd6dc623e571956b29d1bc196869dc5f9a91fa19d03788a87e782818c22bc9c4fe3819fcc8ddf69501a829ebe07271b7ff63a49f124d2daf5854140"
  }
 */
export const TokenCreatedSchema = z.object({
  id: z.string(),
  token: z.string(),
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
    "description": "my first token",
    "duration": "P1Y",
    "permissions": ["permissionsView", "workflowView"]
  }
 */
export const TokenFormSchema = z.object({
  description: z.string().nonempty(),
  duration: ISO8601durationSchema,
  permissions: z.array(z.string()),
});

export const TokenDeletedSchema = z.null();

export type TokenSchemaType = z.infer<typeof TokenSchema>;
export type TokenListSchemaType = z.infer<typeof TokenListSchema>;
export type TokenFormSchemaType = z.infer<typeof TokenFormSchema>;
