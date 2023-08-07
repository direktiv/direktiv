import { z } from "zod";

export const NamespaceSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  name: z.string(),
  oid: z.string(),
});

export const NamespaceListSchema = z.object({
  results: z.array(NamespaceSchema),
});

export const NamespaceCreatedSchema = z.object({
  namespace: NamespaceSchema,
});

// Regex for input validation. This isn't perfect but should be good enough for
// a start. Matches git@hostname:path, where path isn't very restrictive.
export const gitUrlSchema = z
  .string()
  .regex(/^([a-zA-Z0-9.\-_]+@[a-zA-Z0-9.\-_]+:[a-zA-Z0-9.\-_/]+)*$/, {
    message: "format must be git@host:path when using SSH",
  });

// note: in the current API implementation, a mirror is created
// by creating a namespace with this in the payload.
export const MirrorPublicFormSchema = z.object({
  url: z.string().url().nonempty(),
  ref: z.string().nonempty(),
});

// When Token auth is used, token is submitted as "passphrase"
export const MirrorTokenFormSchema = z.object({
  url: z.string().url().nonempty(),
  ref: z.string().nonempty(),
  passphrase: z
    .string()
    .nonempty({ message: "Required when using token auth" }),
});

export const MirrorSshFormSchema = z.object({
  url: gitUrlSchema.nonempty({
    message: "format must be git@host:path when using SSH",
  }),
  ref: z.string().nonempty(),
  passphrase: z.string().optional(),
  privateKey: z.string().nonempty({ message: "Required when using SSH" }),
  publicKey: z.string().nonempty({ message: "Required when using SSH" }),
});

export const MirrorFormSchema = MirrorPublicFormSchema.or(
  MirrorTokenFormSchema
).or(MirrorSshFormSchema);

export type NamespaceListSchemaType = z.infer<typeof NamespaceListSchema>;
export type MirrorPublicFormSchemaType = z.infer<typeof MirrorPublicFormSchema>;
export type MirrorTokenFormSchemaType = z.infer<typeof MirrorTokenFormSchema>;
export type MirrorSshFormSchemaType = z.infer<typeof MirrorSshFormSchema>;
export type MirrorFormSchemaType = z.infer<typeof MirrorFormSchema>;
