import { ButtonType } from "./button";
import { FormType } from "./form";
import { HeadlineType } from "./headline";
import { ModalType } from "./modal";
import { QueryProviderType } from "./queryProvider";
import { TextType } from "./text";
import { TriggerBlocks } from ".";
import { TwoColumnsType } from "./twoColumns";
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

export type TriggerBlocksType = z.infer<typeof TriggerBlocks>;
