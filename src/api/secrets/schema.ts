import { z } from "zod";

export const SecretSchema = z.object({
  name: z.string(),
  initialized: z.boolean(),
});

export const SecretUpdatedSchema = z.object({
  key: z.string(),
  namespace: z.string(),
});

export const SecretListSchema = z.object({
  namespace: z.string(),
  secrets: z.object({
    results: z.array(SecretSchema),
  }),
});

export const SecretDeletedSchema = z.null();

export const SecretFormSchema = z.object({
  name: z.string().nonempty(),
  value: z.string().nonempty(),
});

export type SecretSchemaType = z.infer<typeof SecretSchema>;
export type SecretUpdatedSchemaType = z.infer<typeof SecretUpdatedSchema>;
export type SecretListSchemaType = z.infer<typeof SecretListSchema>;
export type SecretFormSchemaType = z.infer<typeof SecretFormSchema>;
