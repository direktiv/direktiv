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

export const EndpointLoadSchema = z.object({
  "x-direktiv-api": z.literal("endpoint/v2"),
  "x-direktiv-config": z.object({}),
});

export type EndpointLoadSchemaType = z.infer<typeof EndpointLoadSchema>;

export const XDirektivConfigSchema = z.object({
  allow_anonymous: z.boolean().optional(),
  path: z.string().nonempty().optional(),
  timeout: z.number().int().positive().optional(),
  plugins: EndpointsPluginsSchema.optional(),
});

export const EndpointFormSchema = z.object({
  "x-direktiv-api": z.literal("endpoint/v2"),
  "x-direktiv-config": XDirektivConfigSchema,
  ...methodSchemas,
});

export const EndpointSaveSchema = EndpointFormSchema.superRefine(
  (data, ctx) => {
    const hasMethod = routeMethods.some((method) => data[method] !== undefined);
    if (!hasMethod) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "No valid HTTP method available.",
        path: [],
      });
    }

    if (!data["x-direktiv-config"].plugins?.target) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "No target plugin found.",
        path: ["x-direktiv-config", "plugins", "target"],
      });
    }

    if (
      data["x-direktiv-config"].allow_anonymous !== true &&
      !data["x-direktiv-config"].plugins?.auth
    ) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Set allow anonymous to true or add auth plugin.",
        path: ["x-direktiv-config", "allow_anonymous"],
      });
    }
  }
);

export type EndpointFormSchemaType = z.infer<typeof EndpointFormSchema>;
