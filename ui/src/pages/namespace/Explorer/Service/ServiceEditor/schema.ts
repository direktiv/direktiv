import { EnvironementVariableSchema } from "~/api/services/schema/services";
import { z } from "zod";

/**
 * note: scaleOptions should match scale in the ServiceFormSchema,
   but string values are required for the HTML form while numbers
   are stored in the file. */

export const scaleOptions = [
  "0",
  "1",
  "2",
  "3",
  "4",
  "5",
  "6",
  "7",
  "8",
  "9",
] as const;

export const PatchOperations = ["add", "replace", "remove"] as const;

export const PatchOperationSchema = z.enum(PatchOperations);

export const PatchSchema = z.object({
  op: PatchOperationSchema,
  path: z.string().nonempty(),
  value: z.string().nonempty(),
});

export const ServiceFormSchema = z.object({
  direktiv_api: z.literal("service/v1"),
  image: z.string().nonempty().optional(),
  scale: z.number().min(0).lt(10).optional(),
  size: z.string().optional(),
  cmd: z.string().optional(),
  patches: z.array(PatchSchema).nonempty().optional(),
  envs: z.array(EnvironementVariableSchema).nonempty().optional(),
});

export type PatchOperationType = z.infer<typeof PatchOperationSchema>;
export type PatchSchemaType = z.infer<typeof PatchSchema>;
export type ServiceFormSchemaType = z.infer<typeof ServiceFormSchema>;
