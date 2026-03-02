import { z } from "zod";

// ?principal
export const PrincipalSlotSchema = z.literal("?principal");

// ?resource
export const ResourceSlotSchema = z.literal("?resource");

export const SlotSchema = z.union([PrincipalSlotSchema, ResourceSlotSchema]);

export type PrincipalSlotSchemaType = z.infer<typeof PrincipalSlotSchema>;
export type ResourceSlotSchemaType = z.infer<typeof ResourceSlotSchema>;
export type SlotSchemaType = z.infer<typeof SlotSchema>;
