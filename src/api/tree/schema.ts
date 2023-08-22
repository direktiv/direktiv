import { MimeTypeSchema } from "~/pages/namespace/Settings/Variables/MimeTypeSelect";
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

export const ToggleLiveSchema = z.null();

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
    "content-type": z.string(), // same as mimeType
  }),
});

/* needed for validation, but not all properties are editable in the form */
export const WorkflowVariableFormSchema = z.object({
  name: z.string().nonempty(),
  path: z.string().nonempty(),
  content: z.string().nonempty(),
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

/**
 * Example for a mirror-info response
 * {
  "namespace":  "examples",
  "info":  {
    "url":  "https://github.com/direktiv/direktiv-examples",
    "ref":  "main",
    "cron":  "",
    "publicKey":  "",
    "commitId":  "",
    "lastSync":  null,
    "privateKey":  "",
    "passphrase":  ""
  },
  "activities":  {
    "pageInfo":  null,
    "results":  [
      {
        "id":  "29f1c217-2f2a-447d-8730-23f519634755",
        "type":  "init",
        "status":  "complete",
        "createdAt":  "2023-08-04T12:26:18.271385Z",
        "updatedAt":  "2023-08-04T12:26:18.968351Z"
      }
    ]
  }
}
*/

export const MirrorInfoInfoSchema = z.object({
  url: z.string(),
  ref: z.string(),
  lastSync: z.string().or(z.null()),
});

export const MirrorActivitiesSchema = z.object({
  id: z.string(),
  type: z.string(),
  status: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export const MirrorInfoSchema = z.object({
  namespace: z.string(),
  info: MirrorInfoInfoSchema,
  activities: z.object({
    pageInfo: PageinfoSchema.or(z.null()),
    results: z.array(MirrorActivitiesSchema),
  }),
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
export type WorkflowVariableSchemaType = z.infer<typeof WorkflowVariableSchema>;
export type WorkflowVariableFormSchemaType = z.infer<
  typeof WorkflowVariableFormSchema
>;
export type WorkflowVariableCreatedSchemaType = z.infer<
  typeof WorkflowVariableCreatedSchema
>;
