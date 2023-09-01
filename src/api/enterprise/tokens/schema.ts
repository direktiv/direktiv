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
