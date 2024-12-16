import { z } from "zod";

// export const PageSchema = z.object({
//   element: z.any(),
// });

export const LayoutSchema = z.object({
  element: z.string(),
});

export const PageFormSchema = z.object({
  direktiv_api: z.literal("page/v1"),
  path: z.string().nonempty().optional(),
  layout: z.array(LayoutSchema),
});

export type PageFormSchemaType = z.infer<typeof PageFormSchema>;
