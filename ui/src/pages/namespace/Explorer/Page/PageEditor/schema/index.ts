import { z } from "zod";

export const TableKeySchema = z.object({
  header: z.string(),
  cell: z.string(),
});

export type TableKeySchemaType = z.infer<typeof TableKeySchema>;

export const TableSchema = z.array(TableKeySchema).nonempty();

export type TableSchemaType = z.infer<typeof TableSchema>;

export const TextContentSchema = z.object({
  content: z.string(),
});

export type TextContentSchemaType = z.infer<typeof TextContentSchema>;

export const TableContentSchema = z.object({
  dataSourcePath: z.string().optional(),
  dataSourceOutput: z.array(z.string()).optional(),
  content: TableSchema,
});

export type TableContentSchemaType = z.infer<typeof TableContentSchema>;

export const PageElementContentSchema =
  TextContentSchema.or(TableContentSchema);

export type PageElementContentSchemaType = z.infer<
  typeof PageElementContentSchema
>;

export const PageElementSchema = z.object({
  name: z.any(),
  hidden: z.any(),
  content: z.any(),
  preview: z.any(),
});

export type PageElementSchemaType = z.infer<typeof PageElementSchema>;

export const LayoutSchema = z.array(PageElementSchema);

export type LayoutSchemaType = z.infer<typeof LayoutSchema>;

export const PageFormSchema = z.object({
  direktiv_api: z.literal("page/v1"),
  path: z.string().nonempty().optional(),
  layout: z.array(PageElementSchema),
  header: PageElementSchema.optional(),
  footer: PageElementSchema.optional(),
});

export type PageFormSchemaType = z.infer<typeof PageFormSchema>;
