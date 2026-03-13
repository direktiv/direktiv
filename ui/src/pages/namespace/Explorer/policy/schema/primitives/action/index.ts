import {
  AllOperatorSchema,
  EqualOperatorSchema,
  InOperatorSchema,
} from "../shared/operators";
import { EntitySchema } from "../shared/entity";
import type { StrictUnion } from "../../utils/strictUnion";
import { z } from "zod";

type EntityInput = z.input<typeof EntitySchema>;

type ActionType =
  | { op: "All" }
  | { op: "=="; entity: EntityInput }
  | { op: "in"; entity: EntityInput }
  | { op: "in"; entities: EntityInput[] };

// Zod's union input type is too permissive here and can allow object literals
// to mix keys from different action variants. We tighten the input side with a
// strict union so TypeScript rejects extra variant-specific keys, matching the
// runtime `.strict()` schemas.
type ActionInputType = StrictUnion<ActionType>;

// action
const ActionAllSchema = z.object({ op: AllOperatorSchema }).strict();

// action == Action::"readFile"
const ActionEqualSchema = z
  .object({
    op: EqualOperatorSchema,
    entity: EntitySchema,
  })
  .strict();

// action in Action::"readOnly"
const ActionInEntitySchema = z
  .object({
    op: InOperatorSchema,
    entity: EntitySchema,
  })
  .strict();

// action in [Action::"ManageFiles", Action::"readFile"]
const ActionInEntitiesSchema = z
  .object({
    op: InOperatorSchema,
    entities: z.array(EntitySchema),
  })
  .strict();

export const ActionSchema: z.ZodType<
  ActionType,
  z.ZodTypeDef,
  ActionInputType
> = z.union([
  ActionAllSchema,
  ActionEqualSchema,
  ActionInEntitySchema,
  ActionInEntitiesSchema,
]);
