import { inboundPluginTypes } from ".";
import { z } from "zod";

export const AclFormSchema = z.object({
  type: z.literal(inboundPluginTypes.acl.name),
  configuration: z.object({
    allow_groups: z.array(z.string()).optional(),
    deny_groups: z.array(z.string()).optional(),
    allow_tags: z.array(z.string()).optional(),
    deny_tags: z.array(z.string()).optional(),
  }),
});

export type AclFormSchemaType = z.infer<typeof AclFormSchema>;
