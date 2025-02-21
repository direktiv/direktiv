import {
  OperationSchema,
  RouteMethod,
  routeMethods,
} from "~/api/gateway/schema";

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

export const methodSchemas = Array.from(routeMethods).reduce<
  Record<RouteMethod, z.ZodTypeAny>
>(
  (acc, method) => {
    acc[method] = OperationSchema.optional();
    return acc;
  },
  {} as Record<RouteMethod, z.ZodTypeAny>
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
    const hasMethod = Array.from(routeMethods).some(
      (method) => data[method] !== undefined
    );
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

    if (data["x-direktiv-config"].allow_anonymous === false) {
      if (!data["x-direktiv-config"].plugins?.auth) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Auth plugins are missing.",
          path: ["x-direktiv-config", "plugins", "auth"],
        });
      } else if (data["x-direktiv-config"].plugins.auth.length === 0) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Auth plugins array is empty.",
          path: ["x-direktiv-config", "plugins", "auth"],
        });
      }
    }
  }
);

export type EndpointFormSchemaType = z.infer<typeof EndpointFormSchema>;
