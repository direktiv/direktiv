import { z } from "zod";

const EntitySchema = z
  .object({
    type: z.string(),
    id: z.string(),
  })
  .strict();

const ResourceSlotSchema = z.literal("?resource");

// resource
const ResourceAllSchema = z
  .object({
    op: z.literal("All"),
  })
  .strict();

// resource == Folder::"abc"
const ResourceEqualEntitySchema = z
  .object({
    op: z.literal("=="),
    entity: EntitySchema,
  })
  .strict();

// resource == ?resource
const ResourceEqualSlotSchema = z
  .object({
    op: z.literal("=="),
    slot: ResourceSlotSchema,
  })
  .strict();

// resource in Folder::"abc"
const ResourceInEntitySchema = z
  .object({
    op: z.literal("in"),
    entity: EntitySchema,
  })
  .strict();

// resource in ?resource
const ResourceInSlotSchema = z
  .object({
    op: z.literal("in"),
    slot: ResourceSlotSchema,
  })
  .strict();

// resource is Folder in Folder::"Public" / resource is Folder in ?resource
const ResourceIsSchema = z
  .object({
    op: z.literal("is"),
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

export type ResourceSchemaType = z.infer<typeof ResourceSchema>;
