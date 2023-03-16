import { PageinfoSchema } from "../schema";
import { z } from "zod";

const NodeSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  name: z.string(),
  path: z.string(),
  parent: z.string(),
  type: z.string(), // this must be an enum (possible values: directory, ???)
  attributes: z.array(z.string()), // must be specified more
  oid: z.string(),
  readOnly: z.boolean(),
  expandedType: z.string(), // this must be an enum (possible values: git, workflow, directory, ???)
});

export const TreeListSchema = z.object({
  namespace: z.string(),
  node: NodeSchema,
  children: z.object({
    pageinfo: PageinfoSchema,
    results: z.array(NodeSchema),
  }),
});
