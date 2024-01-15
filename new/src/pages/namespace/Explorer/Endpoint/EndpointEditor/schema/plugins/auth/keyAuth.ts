import { authPluginTypes } from ".";
import { z } from "zod";

export const KeyAuthFormSchema = z.object({
  type: z.literal(authPluginTypes.keyAuth),
  configuration: z.object({
    add_username_header: z.boolean(),
    add_tags_header: z.boolean(),
    add_groups_header: z.boolean(),
    key_name: z.string().optional(),
  }),
});

export type KeyAuthFormSchemaType = z.infer<typeof KeyAuthFormSchema>;
