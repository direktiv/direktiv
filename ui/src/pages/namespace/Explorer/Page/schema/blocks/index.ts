import { Button, ButtonType } from "./button";
import { Headline, HeadlineType } from "./headline";
import { Mutation, MutationType } from "../procedures/mutation";
import { Query, QueryType } from "../procedures/query";
import { Text, TextType } from "./text";

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

export const AllBlocks: z.ZodType<AllBlocksType> = z.lazy(() =>
  z.discriminatedUnion("type", [
    Button,
    Form,
    Headline,
    Modal,
    QueryProvider,
    Text,
    TwoColumns,
  ])
);

export const TriggerBlocks = z.discriminatedUnion("type", [Button]);
export type TriggerBlocksType = z.infer<typeof TriggerBlocks>;

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type FormType = {
  type: "form";
  trigger: TriggerBlocksType;
  mutation: MutationType;
  blocks: AllBlocksType[];
};

export const Form = z.object({
  type: z.literal("form"),
  trigger: TriggerBlocks,
  mutation: Mutation,
  blocks: z.array(AllBlocks),
}) satisfies z.ZodType<FormType>;

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type ModalType = {
  type: "modal";
  trigger: TriggerBlocksType;
  blocks: AllBlocksType[];
};

export const Modal = z.object({
  type: z.literal("modal"),
  trigger: TriggerBlocks,
  blocks: z.array(AllBlocks),
}) satisfies z.ZodType<ModalType>;

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type QueryProviderType = {
  type: "queryProvider";
  query: QueryType;
  blocks: AllBlocksType[];
};

export const QueryProvider = z.object({
  type: z.literal("queryProvider"),
  query: Query,
  blocks: z.array(AllBlocks),
}) satisfies z.ZodType<QueryProviderType>;

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type TwoColumnsType = {
  type: "two-columns";
  leftBlocks: AllBlocksType[];
  rightBlocks: AllBlocksType[];
};

export const TwoColumns = z.object({
  type: z.literal("two-columns"),
  leftBlocks: z.array(AllBlocks),
  rightBlocks: z.array(AllBlocks),
}) satisfies z.ZodType<TwoColumnsType>;
