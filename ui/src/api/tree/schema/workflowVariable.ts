import { FileSchema } from "~/api/schema";
import { MimeTypeSchema } from "~/pages/namespace/Settings/Variables/utils";
import { z } from "zod";

/*
 * Example response for a list of variables. Only properties consumed by the 
 * frontend are added to the schemas below.
{
  "namespace":  "foobar",
  "path":  "/file.yaml",
  "variables":  {
    "pageInfo":  null,
    "results":  [
      {
        "name":  "fix-color",
        "createdAt":  "2023-08-17T09:32:34.255765Z",
        "updatedAt":  "2023-08-17T09:32:34.255765Z",
        "checksum":  "",
        "size":  "45",
        "mimeType":  "text/plain"
      },
    ]
  }
}
*/
export const WorkflowVariableSchema = z.object({
  name: z.string(), // identifier
  createdAt: z.string(),
  updatedAt: z.string(),
  mimeType: z.string(), // "application/json"
});

export const WorkflowVariableListSchema = z.object({
  namespace: z.string(),
  path: z.string(), // the workflow identifier
  variables: z.object({
    results: z.array(WorkflowVariableSchema),
  }),
});

export const WorkflowVariableContentSchema = z.object({
  body: z.string(),
  headers: z.object({
    "content-type": z.string().optional(), // same as mimeType
  }),
});

export const WorkflowVariableDownloadSchema = z.object({
  blob: z.instanceof(Blob),
  headers: z.object({
    "content-type": z.string().optional(),
  }),
});

/* needed for validation, but not all properties are editable in the form */
export const WorkflowVariableFormSchema = z.object({
  name: z.string().nonempty(),
  path: z.string().nonempty(),
  content: z.string().nonempty().or(FileSchema),
  mimeType: MimeTypeSchema,
});

/**
 * Example for a workflow variable returned after creating a new record
  {
    "namespace":  "foo",
    "path":  "/bar.yaml",
    "key":  "variable-name",
    "createdAt":  "2023-08-15T13:12:17.432222Z",
    "updatedAt":  "2023-08-15T13:12:17.432222Z",
    "checksum":  "",
    "totalSize":  "10",
    "mimeType":  "application/json"
  }
*/

export const WorkflowVariableCreatedSchema = z.object({
  namespace: z.string(),
  path: z.string(), // workflow path
  key: z.string(), // "name" when GETting the variable
  createdAt: z.string(),
  updatedAt: z.string(),
  mimeType: z.string(), // "application/json"
});

export const WorkflowVariableDeletedSchema = z.null();

export type WorkflowVariableSchemaType = z.infer<typeof WorkflowVariableSchema>;
export type WorkflowVariableFormSchemaType = z.infer<
  typeof WorkflowVariableFormSchema
>;
export type WorkflowVariableCreatedSchemaType = z.infer<
  typeof WorkflowVariableCreatedSchema
>;
