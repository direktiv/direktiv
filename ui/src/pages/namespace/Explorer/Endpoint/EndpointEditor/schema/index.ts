import { OperationSchema, routeMethods } from "~/api/gateway/schema";

import { AuthPluginFormSchema } from "./plugins/auth/schema";
import { InboundPluginFormSchema } from "./plugins/inbound/schema";
import { OutboundPluginFormSchema } from "./plugins/outbound/schema";
import { TargetPluginFormSchema } from "./plugins/target/schema";
import { z } from "zod";

export const EndpointsPluginsSchema = z.object({
  target: TargetPluginFormSchema,
  inbound: z.array(InboundPluginFormSchema).optional(),
  outbound: z.array(OutboundPluginFormSchema).optional(),
  auth: z.array(AuthPluginFormSchema).optional(),
});

export const methodSchemas = routeMethods.reduce<
  Record<(typeof routeMethods)[number], z.ZodTypeAny>
>(
  (acc, method) => {
    acc[method] = OperationSchema.optional();
    return acc;
  },
  {} as Record<(typeof routeMethods)[number], z.ZodTypeAny>
);

export const EndpointFormSchema = z.object({
  "x-direktiv-api": z.literal("endpoint/v2"),
  "x-direktiv-config": z.object({
    allow_anonymous: z.boolean().optional(),
    path: z.string().nonempty().optional(),
    timeout: z.number().int().positive().optional(),
    plugins: EndpointsPluginsSchema.optional(),
  }),
  ...methodSchemas,
});

export type EndpointFormSchemaType = z.infer<typeof EndpointFormSchema>;
