import { z } from "zod";

export const TableSchema = z.array(
  z.object({
    header: z.string(),
    cell: z.string(),
  })
);

export type TableSchemaType = z.infer<typeof TableSchema>;

export const PageElementContentSchema = z.object({
  content: z.string().or(TableSchema),
});

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
