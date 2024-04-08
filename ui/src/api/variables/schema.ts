import { z } from "zod";

const VarTypeSchema = z.enum([
  "namespace-variable",
  "workflow-variable",
  "instance-variable",
]);

/**
 * example:
  {
    "id": "01c9accc-49ab-4acb-a764-551e8ee1eed7",
    "type": "namespace-variable",
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

export const VarContentSchema = z.object({
  data: VarSchema.extend({
    data: z.string(),
  }),
});

export const VarDeletedSchema = z.null();

export const VarCreatedUpdatedSchema = z.object({
  data: VarSchema,
});

export type VarCreatedUpdatedSchemaType = z.infer<
  typeof VarCreatedUpdatedSchema
>;

export const VarFormCreateEditSchema = z.object({
  name: z.string().nonempty(),
  mimeType: z.string().nonempty(),
  data: z.string().nonempty(),
  /**
   * workflowPath is technically not part of the edit payload.
   * The path of a workflow variable can not be changed. But
   * since this is an optional field, it is somewhat compatible
   * with the edit payload and allows us to have one single form/
   * schema for both create and edit.
   */
  workflowPath: z.string().nonempty().optional(),
});

export type VarFormCreateEditSchemaType = z.infer<
  typeof VarFormCreateEditSchema
>;
