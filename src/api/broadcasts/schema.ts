import { ZodBoolean, z } from "zod";

export const BroadcastsSchemaKeys = [
  "directory.create",
  "directory.delete",
  "instance.failed",
  "instance.started",
  "instance.success",
  "instance.variable.create",
  "instance.variable.delete",
  "instance.variable.update",
  "namespace.variable.create",
  "namespace.variable.delete",
  "namespace.variable.update",
  "workflow.create",
  "workflow.delete",
  "workflow.update",
  "workflow.variable.create",
  "workflow.variable.delete",
  "workflow.variable.update",
];

// dynamically create the schema based on the keys
type BroadcastsSchemaKeysType = (typeof BroadcastsSchemaKeys)[number];

const BroadcastsSchemaDefinition: {
  [key in BroadcastsSchemaKeysType]: ZodBoolean;
} = {};

for (const key of BroadcastsSchemaKeys) {
  BroadcastsSchemaDefinition[key] = z.boolean();
}

export const BroadcastsSchema = z.object(BroadcastsSchemaDefinition);

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
