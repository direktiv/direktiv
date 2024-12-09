import { z } from "zod";

const Methods = ["POST", "PUT", "PATCH", "DELETE"] as const;

export const Mutation = z.object({
  method: z.enum(Methods),
  endpoint: z.string().min(1),
  // TODO: finish
});

export type MutationType = z.infer<typeof Mutation>;
