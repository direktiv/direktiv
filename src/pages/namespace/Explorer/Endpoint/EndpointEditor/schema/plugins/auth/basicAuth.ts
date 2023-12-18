import { authPluginTypes } from ".";
import { z } from "zod";

export const BasicAuthFormSchema = z.object({
  type: z.literal(authPluginTypes.basicAuth),
  configuration: z.object({}),
});

export type BasicAuthFormSchemaType = z.infer<typeof BasicAuthFormSchema>;
