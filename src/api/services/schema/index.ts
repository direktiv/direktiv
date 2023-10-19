import { z } from "zod";

export const StatusSchema = z.enum(["True", "False", "Unknown"]);
export const SizeSchema = z.union([z.literal(0), z.literal(1), z.literal(2)]);

export type SizeSchemaType = z.infer<typeof SizeSchema>;
export type StatusSchemaType = z.infer<typeof StatusSchema>;
