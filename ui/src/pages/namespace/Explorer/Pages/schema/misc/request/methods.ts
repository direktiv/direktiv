import { z } from "zod";

const MutationMethods = ["POST", "PUT", "PATCH", "DELETE"] as const;

export const MutationMethod = z.enum(MutationMethods);
