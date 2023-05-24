import { PageinfoSchema } from "../schema";
import { z } from "zod";

export const VarSchema = z.object({
  name: z.string(),
  checksum: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
  size: z.string(),
  mimeType: z.string(),
});

export const VarListSchema = z.object({
  namespace: z.string(),
  variables: z.object({
    pageInfo: PageinfoSchema,
    results: z.array(VarSchema),
  }),
});

export type VarSchemaType = z.infer<typeof VarSchema>;
export type VarListSchemaType = z.infer<typeof VarListSchema>;
