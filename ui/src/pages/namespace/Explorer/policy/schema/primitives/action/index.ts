import { z } from "zod";

const EntitySchema = z
  .object({
    type: z.string(),
    id: z.string(),
  })
  .strict();

// action
const ActionAllSchema = z
  .object({
    op: z.literal("All"),
  })
  .strict();

// action == Action::"readFile"
const ActionEqualSchema = z
  .object({
    op: z.literal("=="),
    entity: EntitySchema,
  })
  .strict();

// action in Action::"readOnly"
const ActionInEntitySchema = z
  .object({
    op: z.literal("in"),
    entity: EntitySchema,
  })
  .strict();

// action in [Action::"ManageFiles", Action::"readFile"]
const ActionInEntitiesSchema = z
  .object({
    op: z.literal("in"),
    entities: z.array(EntitySchema),
  })
  .strict();

export const ActionSchema = z.union([
  ActionAllSchema,
  ActionEqualSchema,
  ActionInEntitySchema,
  ActionInEntitiesSchema,
]);

export type ActionSchemaType = z.infer<typeof ActionSchema>;
