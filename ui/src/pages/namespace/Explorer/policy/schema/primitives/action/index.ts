import {
  AllOperatorSchema,
  EqualOperatorSchema,
  InOperatorSchema,
} from "../shared/operators";
import { EntitySchema } from "../shared/entity";
import { z } from "zod";

// action
const ActionAllSchema = z
  .object({
    op: AllOperatorSchema,
  })
  .strict();

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

export const ActionSchema = z.union([
  ActionAllSchema,
  ActionEqualSchema,
  ActionInEntitySchema,
  ActionInEntitiesSchema,
]);
