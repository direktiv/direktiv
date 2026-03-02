import {
  AllOperatorSchema,
  EqualOperatorSchema,
  InOperatorSchema,
  IsOperatorSchema,
} from "../shared/operators";
import { EntitySchema } from "../shared/entity";
import { ResourceSlotSchema } from "../shared/slot";
import { z } from "zod";

// resource
const ResourceAllSchema = z
  .object({
    op: AllOperatorSchema,
  })
  .strict();

// resource == Folder::"abc"
const ResourceEqualEntitySchema = z
  .object({
    op: EqualOperatorSchema,
    entity: EntitySchema,
  })
  .strict();

// resource == ?resource
const ResourceEqualSlotSchema = z
  .object({
    op: EqualOperatorSchema,
    slot: ResourceSlotSchema,
  })
  .strict();

// resource in Folder::"abc"
const ResourceInEntitySchema = z
  .object({
    op: InOperatorSchema,
    entity: EntitySchema,
  })
  .strict();

// resource in ?resource
const ResourceInSlotSchema = z
  .object({
    op: InOperatorSchema,
    slot: ResourceSlotSchema,
  })
  .strict();

// resource is Folder in Folder::"Public" / resource is Folder in ?resource
const ResourceIsSchema = z
  .object({
    op: IsOperatorSchema,
    entity_type: z.string(),
    in: z
      .union([
        z
          .object({
            entity: EntitySchema,
          })
          .strict(),
        z
          .object({
            slot: ResourceSlotSchema,
          })
          .strict(),
      ])
      .optional(),
  })
  .strict();

export const ResourceSchema = z.union([
  ResourceAllSchema,
  ResourceEqualEntitySchema,
  ResourceEqualSlotSchema,
  ResourceInEntitySchema,
  ResourceInSlotSchema,
  ResourceIsSchema,
]);

type ResourceSchemaType = z.infer<typeof ResourceSchema>;
