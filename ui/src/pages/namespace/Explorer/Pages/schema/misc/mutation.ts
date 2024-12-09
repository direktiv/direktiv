import { z } from "zod";

const Methods = ["POST", "PUT", "PATCH", "DELETE"] as const;

const MutationMethod = z.enum(Methods);

export const Mutation = z.object({
  method: MutationMethod,
  endpoint: z.string().min(1),
  // TODO: finish
});

export type MutationType = z.infer<typeof Mutation>;
