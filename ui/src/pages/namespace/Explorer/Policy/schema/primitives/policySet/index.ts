import { EntitySchema } from "../shared/entity";
import { SlotSchema } from "../shared/slot";
import { z } from "zod";

const PolicySetTemplateLinkSchema = z
  .object({
    templateId: z.string().min(1),
    newId: z.string().min(1),
    values: z.record(SlotSchema, EntitySchema),
  })
  .strict();

export const PolicySetSchema = (policySchema: z.ZodTypeAny) =>
  z
    .object({
      staticPolicies: z.record(policySchema).optional(),
      templates: z.record(policySchema).optional(),
      templateLinks: z.array(PolicySetTemplateLinkSchema).optional(),
    })
    .strict();
