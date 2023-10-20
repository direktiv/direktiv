import { z } from "zod";

export const StatusSchema = z.enum(["True", "False", "Unknown"]);

export type StatusSchemaType = z.infer<typeof StatusSchema>;
