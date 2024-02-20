import { z } from "zod";

export const getFilenameFromPath = (path: string): string => {
  const fileName = path.split("/").pop();
  if (fileName === undefined)
    throw Error(`Filename could not be extracted from ${path}`);
  return fileName;
};

export const getParentFromPath = (path: string): string =>
  path.split("/").slice(0, -1).join("/") || "/";

/* directory example
  {
    path: "/folder",
    type: "directory",
    createdAt: "2024-02-13T15:24:05.856667Z",
    updatedAt: "2024-02-13T15:24:05.856667Z",
  },
*/

/* file example 
  {
    path: "/http.yaml",
    type: "service",
    size: 101,
    mimeType: "application/yaml",
    createdAt: "2024-02-13T09:39:42.78317Z",
    updatedAt: "2024-02-13T09:39:50.137806Z",
  },
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
  children: z.array(BaseFileSchema).nullable(), // only for directories
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

const CreateNodeSchema = z.discriminatedUnion("type", [
  CreateDirectorySchema,
  CreateConsumerSchema,
  CreateEndpointSchema,
  CreateServiceSchema,
  CreateWorkflowSchema,
]);

const RenameNodeSchema = z.object({
  path: z.string(),
});

const UpdateFileSchema = z.object({
  data: z.string(), // base64 encoded file body
});

/**
 * /api/v2/namespaces/:namespace/files/:path
 * 
 * lists the files and directories found under the given path. 
 * "file" lists the item at the current path (this could be a directory).
 * If the returned item is a directory, "paths" will list the items
 * contained in it.
 * 
 * All items are metadata. To request a file's body, make a request to
 * /api/v2/namespaces/:namespace/? TBD
* 
{
  "file": {
    "path": "/",
    "type": "directory",
    "createdAt": "2024-02-12T10:32:58.986418Z",
    "updatedAt": "2024-02-12T10:32:58.986418Z"
  },
  "paths": [
    {
      "path": "/action.yaml",
      "type": "workflow",
      "size": 318,
      "mimeType": "application/direktiv",
      "createdAt": "2024-02-13T08:57:03.81109Z",
      "updatedAt": "2024-02-13T08:57:03.81109Z"
    },
    ..
  ]
}
*/

export const FileListSchema = z.object({
  data: FileSchema,
});

export const PathDeletedSchema = z.null();
export const PathCreatedSchema = z.object({ data: FileSchema });

// data is only returned in the response when it has changed.
export const NodePatchedSchema = z.object({
  data: BaseFileSchema.extend({ data: z.string().optional() }),
});

export type FileSchemaType = z.infer<typeof FileSchema>;
export type BaseFileSchemaType = z.infer<typeof BaseFileSchema>;
export type UpdateFileSchemaType = z.infer<typeof UpdateFileSchema>;
export type RenameFileSchemaType = z.infer<typeof RenameNodeSchema>;

export type CreateFileSchemaType = z.infer<typeof CreateNodeSchema>;
export type FileListSchemaType = z.infer<typeof FileListSchema>;
