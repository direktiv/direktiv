import { PageinfoSchema } from "../schema";
import { z } from "zod";

const NodeSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  name: z.string(),
  path: z.string(),
  parent: z.string(),
  type: z.enum(["directory", "workflow", "file"]),
  attributes: z.array(z.string()), // must be more specified
  oid: z.string(),
  readOnly: z.boolean(),
  expandedType: z.enum(["directory", "workflow", "file", "git"]),
});

const RevisionSchema = z.object({
  createdAt: z.string(),
  hash: z.string(),
  source: z.string(),
  name: z.string(),
});

export const NodeListSchema = z.object({
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

// the revisions scheme in the revisions list only has a subset of the fields
const TrimmedRevisionSchema = z.object({
  name: z.string(),
});

export const RevisionsListSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
  pageInfo: PageinfoSchema,
  results: z.array(TrimmedRevisionSchema),
});

export const TagsListSchema = RevisionsListSchema;

const RouteSchema = z.object({
  ref: z.string(),
  weight: z.number(),
});

export const RouterSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
  live: z.boolean(),
  routes: z
    .array(RouteSchema)
    .refine((routes) => [0, 2].includes(routes.length)),
});

export const FolderCreatedSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
});

export const WorkflowCreatedSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
  revision: RevisionSchema,
});

export const WorkflowStartedSchema = z.object({
  namespace: z.string(),
  instance: z.string(),
});

export const NodeDeletedSchema = z.null();

export const NodeRenameSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
});

export const TagCreatedSchema = z.null();

export const fileNameSchema = z
  .string()
  .regex(/^(([a-z][a-z0-9_\-.]*[a-z0-9])|([a-z]))$/, {
    message:
      "Please use a name that only contains lowercase letters, use - or _ instead of whitespaces.",
  });

export type NodeListSchemaType = z.infer<typeof NodeListSchema>;
export type RevisionsListSchemaType = z.infer<typeof RevisionsListSchema>;
export type TrimmedRevisionSchemaType = z.infer<typeof TrimmedRevisionSchema>;
export type TagsListSchemaType = z.infer<typeof TagsListSchema>;
export type NodeSchemaType = z.infer<typeof NodeSchema>;
export type RouterSchemaType = z.infer<typeof RouterSchema>;
