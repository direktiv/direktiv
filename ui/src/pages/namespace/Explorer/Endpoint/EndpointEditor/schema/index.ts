import { AuthPluginFormSchema } from "./plugins/auth/schema";
import { InboundPluginFormSchema } from "./plugins/inbound/schema";
import { MethodsSchema } from "~/api/gateway/schema";
import { OutboundPluginFormSchema } from "./plugins/outbound/schema";
import { TargetPluginFormSchema } from "./plugins/target/schema";
import { z } from "zod";

export const EndpointFormSchema = z.object({
  direktiv_api: z.literal("endpoint/v1"),
  allow_anonymous: z.boolean().optional(),
  path: z.string().nonempty().optional(),
  timeout: z.number().int().positive().optional(),
  methods: z.array(MethodsSchema).nonempty().optional(),
  plugins: z
    .object({
      target: TargetPluginFormSchema,
      inbound: z.array(InboundPluginFormSchema).optional(),
      outbound: z.array(OutboundPluginFormSchema).optional(),
      auth: z.array(AuthPluginFormSchema).optional(),
    })
    .optional(),
});

export type EndpointFormSchemaType = z.infer<typeof EndpointFormSchema>;
