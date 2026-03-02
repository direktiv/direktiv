import {
  AllOperatorSchema,
  EqualOperatorSchema,
  InOperatorSchema,
  IsOperatorSchema,
} from "../shared/operators";
import { EntitySchema } from "../shared/entity";
import { PrincipalSlotSchema } from "../shared/slot";
import { z } from "zod";

// principal
const PrincipalAllSchema = z
  .object({
    op: AllOperatorSchema,
  })
  .strict();

// principal == User::"alice"
const PrincipalEqualEntitySchema = z
  .object({
    op: EqualOperatorSchema,
    entity: EntitySchema,
  })
  .strict();

// principal == ?principal
const PrincipalEqualSlotSchema = z
  .object({
    op: EqualOperatorSchema,
    slot: PrincipalSlotSchema,
  })
  .strict();

// principal in Group::"Admins"
const PrincipalInEntitySchema = z
  .object({
    op: InOperatorSchema,
    entity: EntitySchema,
  })
  .strict();

// principal in ?principal
const PrincipalInSlotSchema = z
  .object({
    op: InOperatorSchema,
    slot: PrincipalSlotSchema,
  })
  .strict();

// principal is User in Group::"Admins" / principal is User in ?principal
const PrincipalIsSchema = z
  .object({
    op: IsOperatorSchema,
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
