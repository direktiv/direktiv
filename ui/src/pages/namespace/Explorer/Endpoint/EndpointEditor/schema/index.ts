import { AuthPluginFormSchema } from "./plugins/auth/schema";
import { InboundPluginFormSchema } from "./plugins/inbound/schema";
import { MethodsSchema } from "~/api/gateway/schema";
import { OutboundPluginFormSchema } from "./plugins/outbound/schema";
import { TargetPluginFormSchema } from "./plugins/target/schema";
import { forceLeadingSlash } from "~/api/files/utils";
import { z } from "zod";

export const EndpointsPluginsSchema = z.object({
  target: TargetPluginFormSchema,
  inbound: z.array(InboundPluginFormSchema).optional(),
  outbound: z.array(OutboundPluginFormSchema).optional(),
  auth: z.array(AuthPluginFormSchema).optional(),
});

export type EndpointsPluginsSchemaType = z.infer<typeof EndpointsPluginsSchema>;

const processPath = (value: unknown) => {
  // adds leading slash to path when loading file that doesn't have it yet
  if (typeof value !== "string") {
    throw new z.ZodError([
      { message: "Path must be a string", path: ["path"], code: "custom" },
    ]);
  }
  if (value.length === 0) {
    throw new z.ZodError([
      { message: "Path cannot be empty", path: ["path"], code: "custom" },
    ]);
  }

  return forceLeadingSlash(value);
};

export const XDirektivConfigSchema = z.object({
  allow_anonymous: z.boolean().optional(),
  path: z.preprocess((value) => processPath(value), z.string().optional()),
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
