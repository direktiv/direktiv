import { z } from "zod";

export const EffectSchema = z.enum(["permit", "forbid"]);

export type EffectSchemaType = z.infer<typeof EffectSchema>;
