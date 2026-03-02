import { z } from "zod";

export const CedarPolicySchema = z.object({});

export type CedarPolicySchemaType = z.infer<typeof CedarPolicySchema>;
