import { Button, ButtonType } from "./button";
import { Card, CardType } from "./card";
import { Column, ColumnType, Columns, ColumnsType } from "./columns";
import { Dialog, DialogType } from "./dialog";
import { Form, FormType } from "./form";
import { FormCheckbox, FormCheckboxType } from "./form/checkbox";
import { FormInput, FormInputType } from "./form/input";
import { FormSelect, FormSelectType } from "./form/select";
import { FormTextarea, FormTextareaType } from "./form/textarea";
import { Headline, HeadlineType } from "./headline";
import { Image, ImageType } from "./image";
import { Loop, LoopType } from "./loop";
import { QueryProvider, QueryProviderType } from "./queryProvider";
import { Table, TableType } from "./table";
import { Text, TextType } from "./text";

import { ExtractUnionFromSet } from "./utils";
import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the type without updating the schema.
 * The schema needs to get the type input to avoid circular dependencies.
 */

const SimpleBlockUnion = z.discriminatedUnion("type", [
  Button,
  FormCheckbox,
  FormInput,
  FormSelect,
  FormTextarea,
  Headline,
  Image,
  Table,
  Text,
]);

export const ParentBlockUnion = z.discriminatedUnion("type", [
  Card,
  Dialog,
  Form,
  Loop,
  QueryProvider,
  Column,
  Columns,
]);

export type SimpleBlockType =
  | ButtonType
  | FormCheckboxType
  | FormInputType
  | FormSelectType
  | FormTextareaType
  | HeadlineType
  | ImageType
  | TableType
  | TextType;

export type ParentBlockType =
  | CardType
  | DialogType
  | FormType
  | LoopType
  | QueryProviderType
  | ColumnType
  | ColumnsType;

export type BlockType = SimpleBlockType | ParentBlockType;
type BlockTypeUnion = BlockType["type"];

export const Block: z.ZodType<BlockType> = z.lazy(() =>
  z.union([SimpleBlockUnion, ParentBlockUnion])
);

export const AvailableBlockTypeAttributes = z.union([
  z.literal("button"),
  z.literal("headline"),
  z.literal("image"),
  z.literal("table"),
  z.literal("text"),
  z.literal("card"),
  z.literal("dialog"),
  z.literal("form"),
  z.literal("loop"),
  z.literal("query-provider"),
  z.literal("column"),
  z.literal("columns"),
]);

export const TriggerBlock = z.discriminatedUnion("type", [Button]);

export type TriggerBlockType = z.infer<typeof TriggerBlock>;

/* Inline blocks do not need a dialog for creation */
const noFormBlockTypeList = new Set([
  "columns",
  "card",
]) satisfies Set<BlockTypeUnion>;

type noFormBlockTypeUnion = ExtractUnionFromSet<typeof noFormBlockTypeList>;
export type NoFormBlockType = Extract<
  BlockType,
  { type: noFormBlockTypeUnion }
>;

type FormBlockTypeUnion = Exclude<BlockTypeUnion, noFormBlockTypeUnion>;

export type FormBlockType = Extract<BlockType, { type: FormBlockTypeUnion }>;
