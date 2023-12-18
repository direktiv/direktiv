import { inboundPluginTypes } from ".";
import { z } from "zod";

export const AclFormSchema = z.object({
  type: z.literal(inboundPluginTypes.acl),
  configuration: z.object({}),
});

export type AclFormSchemaType = z.infer<typeof AclFormSchema>;
