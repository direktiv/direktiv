import { Button, ButtonType } from "./button";
import { Form, FormType } from "./form";
import { Headline, HeadlineType } from "./headline";
import { Modal, ModalType } from "./modal";
import { QueryProvider, QueryProviderType } from "./queryProvider";
import { Text, TextType } from "./text";
import { TwoColumns, TwoColumnsType } from "./twoColumns";

import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the type without updating the schema.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type AllBlocksType =
  | ButtonType
  | FormType
  | HeadlineType
  | ModalType
  | QueryProviderType
  | TextType
  | TwoColumnsType;

export const AllBlocks: z.ZodType<AllBlocksType> = z.discriminatedUnion(
  "type",
  [Button, Form, Headline, Modal, QueryProvider, Text, TwoColumns]
);

export const TriggerBlocks = z.discriminatedUnion("type", [Button]);

export type TriggerBlocksType = z.infer<typeof TriggerBlocks>;
