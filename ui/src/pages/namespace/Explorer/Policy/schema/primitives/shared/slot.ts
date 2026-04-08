import { z } from "zod";

export const PrincipalSlotSchema = z.literal("?principal");

export const ResourceSlotSchema = z.literal("?resource");

export const SlotSchema = z.union([PrincipalSlotSchema, ResourceSlotSchema]);
