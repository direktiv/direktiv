import { z } from "zod";

// permit(principal, action, resource); / forbid(principal, action, resource);
export const EffectSchema = z.enum(["permit", "forbid"]);

type EffectSchemaType = z.infer<typeof EffectSchema>;
