import { z } from "zod";

const EntitySchema = z
  .object({
    type: z.string(),
    id: z.string(),
  })
  .strict();

const PrincipalSlotSchema = z.literal("?principal");

// principal
const PrincipalAllSchema = z
  .object({
    op: z.literal("All"),
  })
  .strict();

// principal == User::"alice"
const PrincipalEqualEntitySchema = z
  .object({
    op: z.literal("=="),
    entity: EntitySchema,
  })
  .strict();

// principal == ?principal
const PrincipalEqualSlotSchema = z
  .object({
    op: z.literal("=="),
    slot: PrincipalSlotSchema,
  })
  .strict();

// principal in Group::"Admins"
const PrincipalInEntitySchema = z
  .object({
    op: z.literal("in"),
    entity: EntitySchema,
  })
  .strict();

// principal in ?principal
const PrincipalInSlotSchema = z
  .object({
    op: z.literal("in"),
    slot: PrincipalSlotSchema,
  })
  .strict();

// principal is User in Group::"Admins" / principal is User in ?principal
const PrincipalIsSchema = z
  .object({
    op: z.literal("is"),
    entity_type: z.string(),
    in: z
      .union([
        z.object({ entity: EntitySchema }).strict(),
        z.object({ slot: PrincipalSlotSchema }).strict(),
      ])
      .optional(),
  })
  .strict();

export const PrincipalSchema = z.union([
  PrincipalAllSchema,
  PrincipalEqualEntitySchema,
  PrincipalEqualSlotSchema,
  PrincipalInEntitySchema,
  PrincipalInSlotSchema,
  PrincipalIsSchema,
]);

export type PrincipalSchemaType = z.infer<typeof PrincipalSchema>;
