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

const GroupsSchema = z.object({
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
export const GroupslistSchema = z.object({
  groups: z.array(GroupsSchema),
});
