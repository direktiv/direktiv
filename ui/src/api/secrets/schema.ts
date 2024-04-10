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
});

export type SecretsSchemaType = z.infer<typeof SecretSchema>;

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

export const SecretsCreatedUpdatedSchema = z.object({
  data: SecretSchema,
});

export type SecretsCreatedUpdatedSchemaType = z.infer<
  typeof SecretsCreatedUpdatedSchema
>;

export const SecretsFormCreateEditSchema = z.object({
  name: z.string().nonempty(),
  data: z.string().nonempty(),
});

export type SecretFormCreateEditSchemaType = z.infer<
  typeof SecretsFormCreateEditSchema
>;
