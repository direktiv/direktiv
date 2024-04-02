import { z } from "zod";

const VarTypeSchema = z.enum([
  "namespace_variable",
  "workflow_variable",
  "instance_variable",
]);

/**
 * example:
  {
    "id": "01c9accc-49ab-4acb-a764-551e8ee1eed7",
    "type": "namespace_variable",
    "reference": "vars",
    "name": "variable",
    "size": 1,
    "mimeType": "application/json",
    "createdAt": "2024-04-02T06:22:21.766541Z",
    "updatedAt": "2024-04-02T06:22:21.766541Z"
  }
 */
export const VarSchema = z.object({
  id: z.string(),
  type: VarTypeSchema,
  reference: z.string(),
  name: z.string(),
  size: z.number(),
  mimeType: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type VarSchemaType = z.infer<typeof VarSchema>;

/**
 * example:
  {
    "data": [...],
  }
 */
export const VarListSchema = z.object({
  data: z.array(VarSchema),
});

export const VarDeletedSchema = z.null();
