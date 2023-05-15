import { z } from "zod";

export const SecretSchema = z.object({
  name: z.string(),
});

export const SecretListSchema = z.object({
  namespace: z.string(),
  secrets: z.object({
    results: z.array(SecretSchema),
  }),
});

export const SecretDeletedSchema = z.null();

export type SecretSchemaType = z.infer<typeof SecretSchema>;
export type SecretListSchemaType = z.infer<typeof SecretListSchema>;
