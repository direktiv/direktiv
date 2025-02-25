import { AuthPluginFormSchema } from "./plugins/auth/schema";
import { InboundPluginFormSchema } from "./plugins/inbound/schema";
import { MethodsSchema } from "~/api/gateway/schema";
import { OutboundPluginFormSchema } from "./plugins/outbound/schema";
import { TargetPluginFormSchema } from "./plugins/target/schema";
import { z } from "zod";

export const EndpointsPluginsSchema = z.object({
  target: TargetPluginFormSchema.optional(),
  inbound: z.array(InboundPluginFormSchema).optional(),
  outbound: z.array(OutboundPluginFormSchema).optional(),
  auth: z.array(AuthPluginFormSchema).optional(),
});

export type EndpointsPluginsSchemaType = z.infer<typeof EndpointsPluginsSchema>;

export const XDirektivConfigSchema = z.object({
  allow_anonymous: z.boolean().optional(),
  path: z.string().nonempty().optional(),
  timeout: z.number().int().positive().optional(),
  plugins: EndpointsPluginsSchema.optional(),
});

export type XDirektivConfigSchemaType = z.infer<typeof XDirektivConfigSchema>;

export const EndpointFormSchema = z
  .object({
    "x-direktiv-api": z.literal("endpoint/v2"),
    "x-direktiv-config": XDirektivConfigSchema.optional(),
  })
  .merge(MethodsSchema);

export type EndpointFormSchemaType = z.infer<typeof EndpointFormSchema>;
