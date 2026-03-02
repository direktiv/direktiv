import { z } from "zod";

// permit(principal, action, resource); / forbid(principal, action, resource);
export const EffectSchema = z.enum(["permit", "forbid"]);

export type EffectSchemaType = z.infer<typeof EffectSchema>;
