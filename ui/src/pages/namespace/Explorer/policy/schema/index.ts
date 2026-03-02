import { ActionSchema } from "./primitives/action";
import { EffectSchema } from "./primitives/effect";
import { PrincipalSchema } from "./primitives/principal";
import { z } from "zod";

export const CedarPolicySchema = z.object({
  effect: EffectSchema,
  principal: PrincipalSchema,
  action: ActionSchema,
});

export type CedarPolicySchemaType = z.infer<typeof CedarPolicySchema>;
