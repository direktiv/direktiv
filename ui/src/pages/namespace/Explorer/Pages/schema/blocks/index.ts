import { Button, ButtonType } from "./button";
import { Form, FormType } from "./form";
import { Headline, HeadlineType } from "./headline";
import { Modal, ModalType } from "./modal";
import { Text, TextType } from "./text";

import { z } from "zod";

/**
 * ⚠️ NOTE: The AllBlocksType and the allBlocks schema must always be kept in sync
 * to ensure 100% tyoe safety. It is currently possible to extend the AllBlocksType
 * but not implement the schema.
 *
 * The allBlocks need to get the AllBlocksType as a type input to avoid ciuclar
 * dependencies.
 */
type AllBlocksType =
  | HeadlineType
  | ButtonType
  | TextType
  | FormType
  | ModalType;

export const allBlocks: z.ZodType<AllBlocksType> = z.discriminatedUnion(
  "type",
  [Headline, Button, Text, Form, Modal]
);

export const Blocks = {
  all: allBlocks,
  trigger: z.discriminatedUnion("type", [Button]),
};

export type BlocksType = {
  all: z.infer<typeof Blocks.all>;
  trigger: z.infer<typeof Blocks.trigger>;
};
