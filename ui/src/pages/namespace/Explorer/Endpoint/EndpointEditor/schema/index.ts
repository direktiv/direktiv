import { AuthPluginFormSchema } from "./plugins/auth/schema";
import { InboundPluginFormSchema } from "./plugins/inbound/schema";
import { MethodsSchema } from "~/api/gateway/schema";
import { OutboundPluginFormSchema } from "./plugins/outbound/schema";
import { TargetPluginFormSchema } from "./plugins/target/schema";
import { forceLeadingSlash } from "~/api/files/utils";
import { z } from "zod";

const EndpointsPluginsSchema = z.object({
  target: TargetPluginFormSchema,
  inbound: z.array(InboundPluginFormSchema).optional(),
  outbound: z.array(OutboundPluginFormSchema).optional(),
  auth: z.array(AuthPluginFormSchema).optional(),
});

const processPath = (value: unknown) => {
  // adds leading slash to path when loading file that doesn't have it yet
  if (typeof value !== "string") {
    throw new z.ZodError([
      { message: "Path must be a string", path: ["path"], code: "custom" },
    ]);
  }

  return forceLeadingSlash(value);
};

const XDirektivConfigSchema = z.object({
  allow_anonymous: z.boolean().optional(),
  path: z.preprocess((value) => processPath(value), z.string().optional()),
  timeout: z.number().int().positive().optional(),
  plugins: EndpointsPluginsSchema.optional(),
});

export const EndpointFormSchema = z
  .object({
    "x-direktiv-api": z.literal("endpoint/v2"),
    "x-direktiv-config": XDirektivConfigSchema.optional(),
  })
  .merge(MethodsSchema);

export type EndpointFormSchemaType = z.infer<typeof EndpointFormSchema>;
