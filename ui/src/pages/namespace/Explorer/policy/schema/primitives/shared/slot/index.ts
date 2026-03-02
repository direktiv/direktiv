import { z } from "zod";

// ?principal
export const PrincipalSlotSchema = z.literal("?principal");

// ?resource
export const ResourceSlotSchema = z.literal("?resource");

export const SlotSchema = z.union([PrincipalSlotSchema, ResourceSlotSchema]);

type PrincipalSlotSchemaType = z.infer<typeof PrincipalSlotSchema>;
type ResourceSlotSchemaType = z.infer<typeof ResourceSlotSchema>;
type SlotSchemaType = z.infer<typeof SlotSchema>;
