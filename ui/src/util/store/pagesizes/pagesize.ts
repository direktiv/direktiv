import { z } from "zod";

export const pageSizeValue = ["10", "20", "30", "50"] as const;
export const PageSizeValueSchema = z.enum(pageSizeValue);
export type PageSizeValueType = z.infer<typeof PageSizeValueSchema>;
