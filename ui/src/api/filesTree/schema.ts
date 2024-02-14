import { z } from "zod";

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
    mimeType: "application/direktiv",
    createdAt: "2024-02-13T09:39:42.78317Z",
    updatedAt: "2024-02-13T09:39:50.137806Z",
  },
*/

export const direktivNodeTypes = [
  "consumer",
  "directory",
  "endpoint",
  "file",
  "service",
  "workflow",
] as const;

const direktivNodeTypeSchema = z.enum(direktivNodeTypes);

const NodeSchema = z.object({
  type: direktivNodeTypeSchema,
  path: z.string().nonempty(),
  size: z.number().optional(), // only files
  mimeType: z.string().optional(), // only files
  createdAt: z.string(),
  updatedAt: z.string(),
});

/**
 * /api/v2/namespaces/:namespace/files-tree/:path
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

export const PathListSchema = z.object({
  data: z.object({
    file: NodeSchema,
    paths: z.array(NodeSchema),
  }),
});

export type NodeSchemaType = z.infer<typeof NodeSchema>;
export type PathListSchemaType = z.infer<typeof PathListSchema>;
