import { z } from "zod";

export const targetPluginTypes = {
  instantResponse: "instant-response",
  targetFlow: "target-flow",
} as const;

export const InstantResposeFormSchema = z.object({
  type: z.literal(targetPluginTypes.instantResponse),
  configuration: z.object({
    content_type: z.string().nonempty(),
    status_code: z.number().int().positive(),
    status_message: z.string().nonempty(),
  }),
});

export const TargetFlowFormSchema = z.object({
  type: z.literal(targetPluginTypes.targetFlow),
  configuration: z.object({
    flow: z.string().nonempty(),
    content_type: z.string().nonempty(),
    namespace: z.string().nonempty().optional(),
    async: z.boolean().optional(),
  }),
});
