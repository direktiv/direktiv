import { AllBlocks } from "~/pages/namespace/Explorer/Page/poc/schema/blocks";
import z from "zod";

const PathSchema = z.array(z.number());

const AddPayloadSchema = z.object({
  type: z.literal("add"),
  block: AllBlocks,
});

const MovePayloadSchema = z.object({
  type: z.literal("move"),
  block: AllBlocks,
  originPath: PathSchema,
});

export const DragPayloadSchema = z.discriminatedUnion("type", [
  AddPayloadSchema,
  MovePayloadSchema,
]);

export type DragPayloadSchemaType = z.infer<typeof DragPayloadSchema>;

export const DropPayloadSchema = z.object({
  targetPath: PathSchema,
});

export type DropPayloadSchemaType = z.infer<typeof DropPayloadSchema>;

const DragAndDropPayloadSchema = z.object({
  drag: DragPayloadSchema,
  drop: DropPayloadSchema,
});

export type DragAndDropPayloadSchemaType = z.infer<
  typeof DragAndDropPayloadSchema
>;
