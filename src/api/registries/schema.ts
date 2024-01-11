import { z } from "zod";

/**
 * example
  {
    "id": "secret-3dbbd15c3b675a2a9cdb",
    "namespace": "my-namespace",
    "url": "https://domain.com",
    "user": "username" 
  }
 */
export const RegistrySchema = z.object({
  id: z.string(),
  namespace: z.string(),
  url: z.string().url(),
  user: z.string(),
});

export const RegistryFormSchema = z.object({
  url: z.string().url(),
  user: z.string().nonempty(),
  password: z.string().nonempty(),
});

export const RegistryListSchema = z.object({
  data: z.array(RegistrySchema),
});

/**
 * example
  {
    "data": {...}
  }
 */
export const RegistryCreatedSchema = z.object({
  data: RegistrySchema,
});

export const RegistryDeletedSchema = z.null();

export type RegistrySchemaType = z.infer<typeof RegistrySchema>;
export type RegistryCreatedSchemaType = z.infer<typeof RegistryCreatedSchema>;
export type RegistryListSchemaType = z.infer<typeof RegistryListSchema>;
export type RegistryFormSchemaType = z.infer<typeof RegistryFormSchema>;
