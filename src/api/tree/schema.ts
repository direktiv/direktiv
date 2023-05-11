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

const RevisionSchema = z.object({
  createdAt: z.string(),
  hash: z.string(),
  source: z.string(),
  name: z.string(),
});

export const TreeListSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
  children: z
    .object({
      pageInfo: PageinfoSchema,
      results: z.array(NodeSchema),
    })
    .optional(), // not for workflows
  revision: RevisionSchema.optional(), // only for workflows
});

export const RevisionsListSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
  pageInfo: PageinfoSchema,
  results: z.array(
    z.object({
      name: z.string(),
    })
  ),
});

export const TagsListSchema = RevisionsListSchema;

export const TreeFolderCreatedSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
});

export const WorkflowCreatedSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
  revision: RevisionSchema,
});

export const TreeNodeDeletedSchema = z.null();

export const TreeNodeRenameSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
});

export const fileNameSchema = z
  .string()
  .regex(/^(([a-z][a-z0-9_\-.]*[a-z0-9])|([a-z]))$/, {
    message:
      "Please use a name that only contains lowercase letters, use - or _ instead of whitespaces.",
  });

export type TreeListSchemaType = z.infer<typeof TreeListSchema>;
export type RevisionsListSchemaType = z.infer<typeof RevisionsListSchema>;
export type TagsListSchemaType = z.infer<typeof TagsListSchema>;
export type NodeSchemaType = z.infer<typeof NodeSchema>;
