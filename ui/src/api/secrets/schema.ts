import { z } from "zod";

/**
 * example:
  {
    "name": "secret-name",
    "createdAt": "2024-04-02T06:22:21.766541Z",
    "updatedAt": "2024-04-02T06:22:21.766541Z"
  }
 */
export const SecretSchema = z.object({
  name: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
  // TODO: remove the .optional()
  initialized: z.boolean().optional(),
});

export type SecretSchemaType = z.infer<typeof SecretSchema>;

/**
 * example:
  {
    "data": [...],
  }
 */
export const SecretsListSchema = z.object({
  data: z.array(SecretSchema),
});

export const SecretsDeletedSchema = z.null();

export const SecretCreatedUpdatedSchema = z.object({
  data: SecretSchema,
});

export type SecretCreatedUpdatedSchemaType = z.infer<
  typeof SecretCreatedUpdatedSchema
>;

export const SecretFormCreateEditSchema = z.object({
  name: z.string().nonempty(),
  data: z.string().nonempty(),
});

export type SecretFormCreateEditSchemaType = z.infer<
  typeof SecretFormCreateEditSchema
>;
