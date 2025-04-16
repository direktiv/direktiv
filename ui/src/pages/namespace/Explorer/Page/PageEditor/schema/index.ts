import { z } from "zod";

export const pageElementTypes = {
  text: "Text",
  table: "Table",
} as const;

const commonFields = {
  name: z.string().min(1),
  hidden: z.any(),
  preview: z.string(),
};

export const TableKeySchema = z.object({
  header: z.string(),
  cell: z.string(),
});

export type TableKeySchemaType = z.infer<typeof TableKeySchema>;

export const TableSchema = z.array(TableKeySchema);

export type TableSchemaType = z.infer<typeof TableSchema>;

export const TextContentSchema = z.object({
  type: z.literal(pageElementTypes.text),
  content: z.string().nonempty(),
});

export type TextContentSchemaType = z.infer<typeof TextContentSchema>;

export const TableContentSchema = z.object({
  type: z.literal(pageElementTypes.table),
  dataSourcePath: z.string().optional(),
  dataSourceOutput: z.array(z.string()).optional(),
  content: TableSchema,
});

export type TableContentSchemaType = z.infer<typeof TableContentSchema>;

export const PageElementContentSchema = z.discriminatedUnion("type", [
  TextContentSchema,
  TableContentSchema,
]);

export type PageElementContentSchemaType = z.infer<
  typeof PageElementContentSchema
>;

export const PageElementSchema = z.object({
  ...commonFields,
  content: PageElementContentSchema,
});

export type PageElementSchemaType = z.infer<typeof PageElementSchema>;

export const LayoutSchema = z.array(PageElementSchema);

export type LayoutSchemaType = z.infer<typeof LayoutSchema>;

export const PageFormSchema = z.object({
  direktiv_api: z.literal("page/v1"),
  // TODO: [suggestion] is path still needed?
  path: z.string().nonempty().optional(),
  layout: z.array(PageElementSchema),
});

export type PageFormSchemaType = z.infer<typeof PageFormSchema>;
