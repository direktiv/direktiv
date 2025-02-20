import { z } from "zod";

export const permissionTopics = [
  "namespaces",
  "instances",
  "syncs",
  "secrets",
  "variables",
  "files",
  "services",
  "registries",
  "logs",
  "notifications",
  "metrics",
  "events",
] as const;

export type PermissionTopic = (typeof permissionTopics)[number];

const permissionMethods = [
  "POST",
  "GET",
  "DELETE",
  "PATCH",
  "PUT",
  "read",
  "manage",
] as const;

export type PermissionMethod = (typeof permissionMethods)[number];

/**
 * the ui only offers a subset of the methods
 */
export const permissionMethodsAvailableUi = ["manage", "read"] as const;

const PermisionSchema = z.object({
  topic: z.enum(permissionTopics),
  method: z.enum(permissionMethods),
});
export type PermisionSchemaType = z.infer<typeof PermisionSchema>;

export const PermissionsArray = z.array(PermisionSchema);
