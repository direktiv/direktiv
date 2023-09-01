import { z } from "zod";

// Regex for input validation. This isn't perfect but should be good enough for
// a start. Matches git@hostname:path, where path isn't very restrictive.
export const gitUrlSchema = z
  .string()
  .regex(/^([a-zA-Z0-9.\-_]+@[a-zA-Z0-9.\-_]+:[a-zA-Z0-9.\-_/]+)*$/, {
    message: "format must be git@host:path when using SSH",
  });

const PublicValidationSchema = z.object({
  formType: z.literal("public"),
  url: z
    .string()
    .url()
    .nonempty({ message: "invalid url, must be http(s):// format" }),
  ref: z.string().nonempty(),
});

const TokenValidationSchema = z.object({
  formType: z.literal("token"),
  url: z
    .string()
    .url()
    .nonempty({ message: "invalid url, must be http(s):// format" }),
  ref: z.string().nonempty(),
  passphrase: z.string().nonempty("token must not be empty"),
});

const SshValidationSchema = z.object({
  formType: z.literal("ssh"),
  url: gitUrlSchema.nonempty({
    message: "format must be git@host:path when using SSH",
  }),
  ref: z.string().nonempty(),
  passphrase: z.string().optional(),
  publicKey: z.string().nonempty(),
  privateKey: z.string().nonempty(),
});

const KeepTokenValidationSchema = z.object({
  formType: z.literal("keep-token"),
  url: z
    .string()
    .url()
    .nonempty({ message: "invalid url, must be http(s):// format" }),
  ref: z.string().nonempty(),
});

const KeepSshValidationSchema = z.object({
  formType: z.literal("keep-ssh"),
  url: gitUrlSchema.nonempty({
    message: "format must be git@host:path when using SSH",
  }),
  ref: z.string().nonempty(),
});

export const MirrorValidationSchema = z.discriminatedUnion("formType", [
  PublicValidationSchema,
  SshValidationSchema,
  TokenValidationSchema,
  KeepSshValidationSchema,
  KeepTokenValidationSchema,
]);
