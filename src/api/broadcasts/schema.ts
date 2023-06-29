import { z } from "zod";

export const BroadcastsSchema = z.object({
  "directory.create": z.boolean(),
  "directory.delete": z.boolean(),
  "instance.failed": z.boolean(),
  "instance.started": z.boolean(),
  "instance.success": z.boolean(),
  "instance.variable.create": z.boolean(),
  "instance.variable.delete": z.boolean(),
  "instance.variable.update": z.boolean(),
  "namespace.variable.create": z.boolean(),
  "namespace.variable.delete": z.boolean(),
  "namespace.variable.update": z.boolean(),
  "workflow.create": z.boolean(),
  "workflow.delete": z.boolean(),
  "workflow.update": z.boolean(),
  "workflow.variable.create": z.boolean(),
  "workflow.variable.delete": z.boolean(),
  "workflow.variable.update": z.boolean(),
});

export const BroadcastsResponseSchema = z.object({
  broadcast: BroadcastsSchema,
});

export const BroadcastsPatchSchema = z.object({
  broadcast: BroadcastsSchema.partial(),
});

export type BroadcastsSchemaType = z.infer<typeof BroadcastsSchema>;
export type BroadcastsResponseSchemaType = z.infer<
  typeof BroadcastsResponseSchema
>;
export type BroadcastsPatchSchemaType = z.infer<typeof BroadcastsPatchSchema>;
