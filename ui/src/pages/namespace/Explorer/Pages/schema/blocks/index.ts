import { Button, ButtonType } from "./button";
import { Headline, HeadlineType } from "./headline";
import { Modal, ModalType } from "./modal";
import { Text, TextType } from "./text";

import { z } from "zod";

/**
 * ⚠️ NOTE: The AllBlocks unions and the AllBlocks schema must always be kept in sync to
 * ensure 100% tyoe safety. It is currently possible to extend the AllBlocks union type
 * but not implement the schema for the new type.
 *
 * The allBlocks need to get the AllBlocksType union type as input to avoid ciuclar
 * dependencies.
 */
type AllBlocksType = HeadlineType | ButtonType | TextType | ModalType;

export const allBlocks: z.ZodType<AllBlocksType> = z.discriminatedUnion(
  "type",
  [Headline, Button, Text, Modal]
);

export const Blocks = {
  all: allBlocks,
  trigger: z.discriminatedUnion("type", [Button]),
};

export type BlocksType = {
  all: z.infer<typeof Blocks.all>;
  trigger: z.infer<typeof Blocks.trigger>;
};
