import { FileSchema } from "../schema";
import { MimeTypeSchema } from "~/pages/namespace/Settings/Variables/MimeTypeSelect";
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

export const VarContentSchema = z.object({
  body: z.string(),
  headers: z.object({
    "content-type": z.string().optional(),
  }),
});

export const VarDownloadSchema = z.object({
  blob: z.instanceof(Blob),
  headers: z.object({
    "content-type": z.string().optional(),
  }),
});

export const VarListSchema = z.object({
  namespace: z.string(),
  variables: z.object({
    results: z.array(VarSchema),
  }),
});

export const VarFormSchema = z.object({
  name: z.string().nonempty(),
  content: z.string().nonempty().or(FileSchema),
  mimeType: MimeTypeSchema,
});

export type VarSchemaType = z.infer<typeof VarSchema>;
export type VarUpdatedSchemaType = z.infer<typeof VarUpdatedSchema>;
export type VarContentSchemaType = z.infer<typeof VarContentSchema>;
export type VarFormSchemaType = z.infer<typeof VarFormSchema>;
export type VarListSchemaType = z.infer<typeof VarListSchema>;
