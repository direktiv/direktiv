import { z } from "zod";

export const RegistrySchema = z.object({
  id: z.string(),
  name: z.string().url(),
  user: z.string(),
});

export const RegistryFormSchema = z.object({
  url: z.string().url(),
  user: z.string().nonempty(),
  password: z.string().nonempty(),
});

// this is the format required when POSTing a new registry
export const RegistryPostSchema = z.object({
  data: z.string(), // format: user:pwd
  reg: z.string().url(), // this is the url
});

export const RegistryListSchema = z.object({
  registries: z.array(RegistrySchema),
});

export const RegistryCreatedSchema = z.null();

export const RegistryDeletedSchema = z.null();

export type RegistrySchemaType = z.infer<typeof RegistrySchema>;
export type RegistryCreatedSchemaType = z.infer<typeof RegistryCreatedSchema>;
export type RegistryListSchemaType = z.infer<typeof RegistryListSchema>;
export type RegistryFormSchemaType = z.infer<typeof RegistryFormSchema>;
