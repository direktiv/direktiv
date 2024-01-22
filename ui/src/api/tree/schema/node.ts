import { PageinfoSchema } from "../../schema";
import { z } from "zod";

const NodeSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  name: z.string(),
  path: z.string(),
  parent: z.string(),
  type: z.enum([
    "directory",
    "workflow",
    "file",
    "service",
    "endpoint",
    "consumer",
  ]),
  attributes: z.array(z.string()), // must be more specified
  oid: z.string(),
  readOnly: z.boolean(),
  expandedType: z.enum(["directory", "workflow", "file", "git"]),
  mimeType: z.string(),
});

export const NodeListSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
  source: z.string().optional(), // only for workflows
  children: z
    .object({
      pageInfo: PageinfoSchema,
      results: z.array(NodeSchema),
    })
    .optional(), // not for workflows
});

export const FolderCreatedSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
});

export const WorkflowCreatedSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
  source: z.string(),
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

export const fileNameSchema = z
  .string()
  .regex(/^(([a-z][a-z0-9_\-.]*[a-z0-9])|([a-z]))$/, {
    message:
      "Please use a name that only contains lowercase letters, use - or _ instead of whitespaces.",
  });

export type NodeListSchemaType = z.infer<typeof NodeListSchema>;
export type NodeSchemaType = z.infer<typeof NodeSchema>;
