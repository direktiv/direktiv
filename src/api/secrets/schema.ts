import { z } from "zod";

export const SecretSchema = z.object({
  name: z.string(),
});

export const SecretCreatedSchema = z.object({
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

export type SecretSchemaType = z.infer<typeof SecretSchema>;
export type SecretCreatedSchemaType = z.infer<typeof SecretCreatedSchema>;
export type SecretListSchemaType = z.infer<typeof SecretListSchema>;
