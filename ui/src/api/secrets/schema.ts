import { z } from "zod";

/**
 * example:
  {
    "name": "secret-name",
    "createdAt": "2024-04-02T06:22:21.766541Z",
    "updatedAt": "2024-04-02T06:22:21.766541Z"
  }
 */
const SecretSchema = z.object({
  name: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
  initialized: z.boolean(),
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

export const SecretFormCreateEditSchema = z.object({
  name: z
    .string()
    .nonempty()
    .regex(/^[a-z-]+$/, {
      message: "Only lowercase letters and dashes are allowed",
    }),
  data: z.string().nonempty(),
});

export type SecretFormCreateEditSchemaType = z.infer<
  typeof SecretFormCreateEditSchema
>;
