import { ActionSchema } from "./primitives/action";
import { AnnotationsSchema } from "./primitives/annotations";
import { ConditionsSchema } from "./primitives/conditions";
import { EffectSchema } from "./primitives/effect";
import { PolicySetSchema } from "./primitives/policySet";
import { PrincipalSchema } from "./primitives/principal";
import { ResourceSchema } from "./primitives/resource";
import { z } from "zod";

export const CedarPolicySchema = z.object({
  effect: EffectSchema,
  principal: PrincipalSchema,
  action: ActionSchema,
  resource: ResourceSchema,
  conditions: ConditionsSchema,
  annotations: AnnotationsSchema.optional(),
});

export const CedarPolicySetSchema = PolicySetSchema(CedarPolicySchema);

export type CedarPolicySchemaType = z.infer<typeof CedarPolicySchema>;
export type CedarPolicySetSchemaType = z.infer<typeof CedarPolicySetSchema>;
