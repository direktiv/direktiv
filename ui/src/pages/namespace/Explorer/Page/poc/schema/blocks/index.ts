import { Button, ButtonType } from "./button";
import { Card, CardType } from "./card";
import { Column, ColumnType, Columns, ColumnsType } from "./columns";
import { Dialog, DialogType } from "./dialog";
import { Form, FormType } from "./form";
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

const SimpleBlocksUnion = z.discriminatedUnion("type", [
  Button,
  Headline,
  Image,
  Table,
  Text,
]);

export const ParentBlocksUnion = z.discriminatedUnion("type", [
  Card,
  Dialog,
  Form,
  Loop,
  QueryProvider,
  Column,
  Columns,
]);

export type SimpleBlocksType =
  | ButtonType
  | HeadlineType
  | ImageType
  | TableType
  | TextType;

export type ParentBlocksType =
  | CardType
  | DialogType
  | FormType
  | LoopType
  | QueryProviderType
  | ColumnType
  | ColumnsType;

export type AllBlocksType = SimpleBlocksType | ParentBlocksType;
type AllBlocksTypeUnion = AllBlocksType["type"];

export const AllBlocks: z.ZodType<AllBlocksType> = z.lazy(() =>
  z.union([SimpleBlocksUnion, ParentBlocksUnion])
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

export const TriggerBlocks = z.discriminatedUnion("type", [Button]);

export type TriggerBlocksType = z.infer<typeof TriggerBlocks>;

/* Inline blocks do not need a dialog for creation */
const noFormBlocksTypeList = new Set([
  "columns",
  "card",
]) satisfies Set<AllBlocksTypeUnion>;

type noFormBlocksTypeUnion = ExtractUnionFromSet<typeof noFormBlocksTypeList>;
export type NoFormBlocksType = Extract<
  AllBlocksType,
  { type: noFormBlocksTypeUnion }
>;

type FormBlocksTypeUnion = Exclude<AllBlocksTypeUnion, noFormBlocksTypeUnion>;

export type FormBlocksType = Extract<
  AllBlocksType,
  { type: FormBlocksTypeUnion }
>;
