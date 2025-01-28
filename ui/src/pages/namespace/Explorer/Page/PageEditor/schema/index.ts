import { z } from "zod";

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
});

export type PageFormSchemaType = z.infer<typeof PageFormSchema>;
