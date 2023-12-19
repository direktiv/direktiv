import { authPluginTypes } from ".";
import { z } from "zod";

export const KeyAuthFormSchema = z.object({
  type: z.literal(authPluginTypes.keyAuth),
  configuration: z.object({}),
});

export type KeyAuthFormSchemaType = z.infer<typeof KeyAuthFormSchema>;
