import { PermissionsArray } from "../schema";
import { z } from "zod";

/**
 * example:
 * 
{
    "name": "foo1",
    "description": "foo1 description",
    "oidcGroups": ["foo1_g1", "foo1_g2"],
    "permissions": [
      {
        "topic": "secrets",
        "method": "read"
      },
      {
        "topic": "variables",
        "method": "manage"
      }
    ],
    "createdAt": "2024-02-05T12:00:00Z",
    "updatedAt": "2024-02-05T12:00:00Z"
  }
}
 */

const RoleSchema = z.object({
  name: z.string(),
  description: z.string(),
  oidcGroups: z.array(z.string()),
  permissions: PermissionsArray,
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
export const RolesListSchema = z.object({
  data: z.array(RoleSchema),
});

/**
 * example:
 * 
  {
    "data": {...}
  }
 */
export const RoleCreatedEditedSchema = z.object({
  data: RoleSchema,
});

/**
 * example
 * 
  {
    "name": "role name,
    "description": "role description",
    "oidcGroups": ["foo1_g1", "foo1_g2"],
    "permissions": [
      { "topic": "namespace", "method": "read" },
      { "topic": "files", "method": "manage" }
    ]
  }
 */
export const RoleFormSchema = z.object({
  name: z.string().nonempty(),
  description: z.string(),
  oidcGroups: z.array(z.string()),
  permissions: PermissionsArray,
});

export const RoleDeletedSchema = z.null();

export type RoleSchemaType = z.infer<typeof RoleSchema>;
export type RolesListSchemaType = z.infer<typeof RolesListSchema>;
export type RoleFormSchemaType = z.infer<typeof RoleFormSchema>;
