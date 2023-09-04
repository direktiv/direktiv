import { z } from "zod";

/**
 * example:
 * 
  {
    "id": "a9ad7eb6-f33b-4094-b4bb-19aaa3462611",
    "group": "mygroup",
    "description": "desc1",
    "permissions": [
      "workflowView",
      "permissionsView"
    ]
  }
 */

const GroupSchema = z.object({
  id: z.string(),
  group: z.string(),
  description: z.string(),
  permissions: z.array(z.string()),
});

/**
 * example:
 * 
  {
    "groups": [...]
  }
 */
export const GroupsListSchema = z.object({
  groups: z.array(GroupSchema),
});

/**
 * example:
 * 
  { "id": "e18d5300-7b16-4d77-afb2-d6c969978895" }
 */
export const GroupCreatedEditedSchema = z.object({
  id: z.string(),
});

/**
 * example
 * 
  {
    "description": "desc1",
    "group": "mygroup",
    "permissions": ["permissionsView", "workflowView"]
  }
 */
export const GroupFormSchema = z.object({
  description: z.string(),
  group: z.string(),
  permissions: z.array(z.string()),
});

export const GroupDeletedSchema = z.null();

export type GroupSchemaType = z.infer<typeof GroupSchema>;
export type GroupsListSchemaType = z.infer<typeof GroupsListSchema>;
export type GroupFormSchemaType = z.infer<typeof GroupFormSchema>;
