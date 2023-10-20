import { z } from "zod";

export const StatusSchema = z.enum(["True", "False", "Unknown"]);
export const SizeSchema = z.enum(["small", "medium", "large"]);

export type SizeSchemaType = z.infer<typeof SizeSchema>;
export type StatusSchemaType = z.infer<typeof StatusSchema>;
