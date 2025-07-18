import { AllBlocks } from "~/pages/namespace/Explorer/Page/poc/schema/blocks";
import z from "zod";

const AddPayloadSchema = z.object({
  type: z.literal("add"),
  block: AllBlocks,
});

const MovePayloadSchema = z.object({
  type: z.literal("move"),
  block: AllBlocks,
  originPath: z.array(z.number()),
});

export const PayloadSchema = z.discriminatedUnion("type", [
  AddPayloadSchema,
  MovePayloadSchema,
]);

export type PayloadSchemaType = z.infer<typeof PayloadSchema>;
