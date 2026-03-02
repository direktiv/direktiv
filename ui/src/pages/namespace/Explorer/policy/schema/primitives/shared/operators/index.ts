import { z } from "zod";

export const AllOperatorSchema = z.literal("All");
export const EqualOperatorSchema = z.literal("==");
export const InOperatorSchema = z.literal("in");
export const IsOperatorSchema = z.literal("is");

export type AllOperatorSchemaType = z.infer<typeof AllOperatorSchema>;
export type EqualOperatorSchemaType = z.infer<typeof EqualOperatorSchema>;
export type InOperatorSchemaType = z.infer<typeof InOperatorSchema>;
export type IsOperatorSchemaType = z.infer<typeof IsOperatorSchema>;
