import { TemplateString } from "../../primitives/templateString";
import { z } from "zod";

export const TableColumn = z.object({
  type: z.literal("table-column"),
  label: TemplateString,
  content: TemplateString,
});

export type TableColumnType = z.infer<typeof TableColumn>;
