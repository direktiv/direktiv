import { z } from "zod";

export const AllOperatorSchema = z.literal("All");
export const EqualOperatorSchema = z.literal("==");
export const InOperatorSchema = z.literal("in");
export const IsOperatorSchema = z.literal("is");

type AllOperatorSchemaType = z.infer<typeof AllOperatorSchema>;
type EqualOperatorSchemaType = z.infer<typeof EqualOperatorSchema>;
type InOperatorSchemaType = z.infer<typeof InOperatorSchema>;
type IsOperatorSchemaType = z.infer<typeof IsOperatorSchema>;
