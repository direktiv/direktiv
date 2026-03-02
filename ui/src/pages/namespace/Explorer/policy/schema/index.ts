import { EffectSchema } from "./primitives/effect";
import { PrincipalSchema } from "./primitives/principal";
import { z } from "zod";

export const CedarPolicySchema = z.object({
  effect: EffectSchema,
  principal: PrincipalSchema,
});

export type CedarPolicySchemaType = z.infer<typeof CedarPolicySchema>;
