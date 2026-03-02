import { EffectSchema } from "./primitives/effect";
import { z } from "zod";

export const CedarPolicySchema = z.object({
  effect: EffectSchema,
});

export type CedarPolicySchemaType = z.infer<typeof CedarPolicySchema>;
