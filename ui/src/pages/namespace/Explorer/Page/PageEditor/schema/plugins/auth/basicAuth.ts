import { authPluginTypes } from ".";
import { z } from "zod";

export const BasicAuthFormSchema = z.object({
  type: z.literal(authPluginTypes.basicAuth.name),
  configuration: z.object({
    add_username_header: z.boolean(),
    add_tags_header: z.boolean(),
    add_groups_header: z.boolean(),
  }),
});

export type BasicAuthFormSchemaType = z.infer<typeof BasicAuthFormSchema>;
