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

export const VarUpdatedSchema = z.object({
  namespace: z.string(),
  key: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
  checksum: z.string(),
  totalSize: z.string(),
  mimeType: z.string(),
});

export const VarDeletedSchema = z.null();

export const VarContentSchema = z.string();

export const VarListSchema = z.object({
  namespace: z.string(),
  variables: z.object({
    pageInfo: PageinfoSchema,
    results: z.array(VarSchema),
  }),
});

export const VarFormSchema = z.object({
  name: z.string(),
  content: z.string(),
});

export type VarSchemaType = z.infer<typeof VarSchema>;
export type VarUpdatedSchemaType = z.infer<typeof VarUpdatedSchema>;
export type VarContentSchemaType = z.infer<typeof VarContentSchema>;
export type VarFormSchemaType = z.infer<typeof VarFormSchema>;
export type VarListSchemaType = z.infer<typeof VarListSchema>;
