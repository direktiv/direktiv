import { PageinfoSchema } from "../schema";
import { z } from "zod";

const NodeSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  name: z.string(),
  path: z.string(),
  parent: z.string(),
  type: z.enum(["directory", "workflow"]),
  attributes: z.array(z.string()), // must be more specified
  oid: z.string(),
  readOnly: z.boolean(),
  expandedType: z.enum(["directory", "workflow", "git"]),
});

export const TreeListSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
  children: z
    .object({
      pageInfo: PageinfoSchema,
      results: z.array(NodeSchema),
    })
    .optional(),
});

export const TreeFolderCreatedSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
});

export const TreeNodeDeletedSchema = z.null();

export const fileNameSchema = z
  .string()
  .regex(/^(([a-z][a-z0-9_\-.]*[a-z0-9])|([a-z]))$/, {
    message:
      "Please use a folder name that starts with a lowercase letter, use - or _ instead of whitespaces.",
  });

export type TreeListSchemaType = z.infer<typeof TreeListSchema>;
export type NodeSchemaType = z.infer<typeof NodeSchema>;
