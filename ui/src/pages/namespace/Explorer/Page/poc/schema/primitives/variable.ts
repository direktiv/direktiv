import { z } from "zod";

export const Variable = z.string().min(1);

export type VariableType = z.infer<typeof Variable>;
