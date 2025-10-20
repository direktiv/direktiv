import { WorkflowValidationSchema } from "../validate/schema";
import { z } from "zod";

/**
 * /api/v2/namespaces/:namespace/files/:path
 * 
 * lists the files and directories found under the given path. 
 * "file" lists the item at the current path (this could be a directory).
 * If the returned item is a directory, "paths" will list the items
 * contained in it.
 * 
 * Example response for directory:
  {
    "data": {
      "path": "/",
      "type": "directory",
      "createdAt": "2024-02-12T10:32:58.986418Z",
      "updatedAt": "2024-02-12T10:32:58.986418Z",
      "children": [
        {
          "path": "/aaa",
          "type": "directory",
          "createdAt": "2024-02-13T15:24:05.856667Z",
          "updatedAt": "2024-02-15T16:33:13.79461Z"
        },
        {
          "path": "/aaaa.yaml",
          "type": "service",
          "size": 212,
          "mimeType": "application/direktiv",
          "createdAt": "2024-02-13T10:39:57.730916Z",
          "updatedAt": "2024-02-15T16:33:13.79461Z"
        },   
      ]  
    }
 *
 * Example response for file:
 * 
  {
    "data": {
      "path": "/aaaa.yaml",
      "type": "service",
      "data": "base64-encoded-string",
      "size": 212,
      "mimeType": "application/direktiv",
      "createdAt": "2024-02-13T10:39:57.730916Z",
      "updatedAt": "2024-02-15T16:33:13.79461Z",
      "children": null
    }
  }
 */

const fileTypes = [
  "consumer",
  "directory",
  "endpoint",
  "file",
  "service",
  "workflow",
  "gateway",
] as const;

const FileTypeSchema = z.enum(fileTypes);

/* All filesystem records (including "directories") have these properties. */
const BaseFileSchema = z.object({
  type: FileTypeSchema,
  path: z.string().nonempty(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

/* When a specific single file is returned, it also has these properties */
const FileSchema = BaseFileSchema.extend({
  type: FileTypeSchema.exclude(["directory"]),
  size: z.number(),
  mimeType: z.string(),
  data: z.string(),
});

/* Additional properties exist on files in "children", but aren't currently used. */
const DirectorySchema = BaseFileSchema.extend({
  type: z.literal("directory"),
  children: z.array(BaseFileSchema).optional(),
});

const CreateDirectorySchema = z.object({
  type: z.literal("directory"),
  name: z.string().nonempty(),
});

const CreateYamlFileSchema = z.object({
  type: z.enum(["consumer", "endpoint", "service", "workflow", "gateway"]),
  name: z.string().nonempty(),
  mimeType: z.literal("application/yaml"),
  data: z.string(), // base64 encoded file body
});

const CreateTSFileSchema = z.object({
  name: z.string().nonempty(),
  mimeType: z.literal("application/x-typescript"),
  data: z.string(), // base64 encoded file body
});

const CreateConsumerSchema = CreateYamlFileSchema.extend({
  type: z.literal("consumer"),
});

const CreateEndpointSchema = CreateYamlFileSchema.extend({
  type: z.literal("endpoint"),
});

const CreateServiceSchema = CreateYamlFileSchema.extend({
  type: z.literal("service"),
});

const CreateWorkflowSchema = CreateTSFileSchema.extend({
  type: z.literal("workflow"),
});

const CreateGatewaySchema = CreateYamlFileSchema.extend({
  type: z.literal("gateway"),
});

export const CreateFileSchema = z.discriminatedUnion("type", [
  CreateDirectorySchema,
  CreateConsumerSchema,
  CreateEndpointSchema,
  CreateServiceSchema,
  CreateWorkflowSchema,
  CreateGatewaySchema,
]);

const _RenameFileSchema = z.object({
  path: z.string(),
});

const _UpdateFileSchema = z.object({
  data: z.string(), // base64 encoded file body
});

export const FileListSchema = z.object({
  data: DirectorySchema.or(FileSchema),
});

export const FileDeletedSchema = z.null();

/**
 * expected response for
 * POST /api/v2/namespaces/:namespace/files/
 *
 * The actual response contains more data, but since we do not use
 * it, we do not bother defining it here.
 * data is only present in the response when it has changed.
 */
export const SaveFileResponseSchema = z.object({
  data: BaseFileSchema.extend({
    data: z.string().optional(),
    errors: WorkflowValidationSchema,
  }),
});

export const FileNameSchema = z
  .string()
  .regex(/^(([a-z][a-z0-9_\-.]*[a-z0-9])|([a-z]))$/, {
    message:
      "Please use a name that only contains lowercase letters, use - or _ instead of whitespaces.",
  });

export type BaseFileSchemaType = z.infer<typeof BaseFileSchema>;
export type FileSchemaType = z.infer<typeof FileSchema>;
export type UpdateFileSchemaType = z.infer<typeof _UpdateFileSchema>;
export type RenameFileSchemaType = z.infer<typeof _RenameFileSchema>;
export type CreateFileSchemaType = z.infer<typeof CreateFileSchema>;
export type SaveFileResponseSchemaType = z.infer<typeof SaveFileResponseSchema>;
