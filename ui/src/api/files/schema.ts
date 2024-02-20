import { z } from "zod";

export const getFilenameFromPath = (path: string): string => {
  const fileName = path.split("/").pop();
  if (fileName === undefined)
    throw Error(`Filename could not be extracted from ${path}`);
  return fileName;
};

export const getParentFromPath = (path: string): string =>
  path.split("/").slice(0, -1).join("/") || "/";

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

export const direktivFileTypes = [
  "consumer",
  "directory",
  "endpoint",
  "file",
  "service",
  "workflow",
] as const;

const direktivFileTypeSchema = z.enum(direktivFileTypes);

const BaseFileSchema = z.object({
  type: direktivFileTypeSchema,
  path: z.string().nonempty(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

const FileSchema = z.object({
  type: direktivFileTypeSchema,
  path: z.string().nonempty(),
  createdAt: z.string(),
  updatedAt: z.string(),
  size: z.number().optional(), // not for directories
  mimeType: z.string().optional(), // not for directories
  data: z.string().optional(), // not for directories
  children: z.array(BaseFileSchema).nullable().optional(), // only for directories
});

const CreateDirectorySchema = z.object({
  type: z.literal("directory"),
  name: z.string().nonempty(),
});

const CreateYamlFileSchema = z.object({
  type: z.enum(["consumer", "endpoint", "service", "workflow"]),
  name: z.string().nonempty(),
  mimeType: z.literal("application/yaml"),
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

const CreateWorkflowSchema = CreateYamlFileSchema.extend({
  type: z.literal("workflow"),
});

const CreateFileSchema = z.discriminatedUnion("type", [
  CreateDirectorySchema,
  CreateConsumerSchema,
  CreateEndpointSchema,
  CreateServiceSchema,
  CreateWorkflowSchema,
]);

const RenameFileSchema = z.object({
  path: z.string(),
});

const UpdateFileSchema = z.object({
  data: z.string(), // base64 encoded file body
});

export const FileListSchema = z.object({
  data: FileSchema,
});

export const FileDeletedSchema = z.null();
export const FileCreatedSchema = z.object({ data: FileSchema });

// data is only returned in the response when it has changed.
export const FilePatchedSchema = z.object({
  data: BaseFileSchema.extend({ data: z.string().optional() }),
});

export type FileSchemaType = z.infer<typeof FileSchema>;
export type BaseFileSchemaType = z.infer<typeof BaseFileSchema>;
export type UpdateFileSchemaType = z.infer<typeof UpdateFileSchema>;
export type RenameFileSchemaType = z.infer<typeof RenameFileSchema>;

export type CreateFileSchemaType = z.infer<typeof CreateFileSchema>;
export type FileListSchemaType = z.infer<typeof FileListSchema>;
