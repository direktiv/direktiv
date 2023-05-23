import { z } from "zod";

export const RegistrySchema = z.object({
  id: z.string(),
  name: z.string().url(),
  user: z.string(),
});

// this is the format required when POSTing a new registry
export const RegistryPostSchema = z.object({
  data: z.string(), // format: user:pwd
  reg: z.string(), // this is the url
});

// not sure how this is supposed to look
export const RegistryCreatedSchema = z.object({});

export const RegistryListSchema = z.object({
  registries: z.array(RegistrySchema),
});

export const RegistryDeletedSchema = z.null();

export type RegistrySchemaType = z.infer<typeof RegistrySchema>;
export type RegistryCreatedSchemaType = z.infer<typeof RegistryCreatedSchema>;
export type RegistryListSchemaType = z.infer<typeof RegistryListSchema>;
