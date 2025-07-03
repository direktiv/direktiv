import { Button, ButtonType } from "./button";
import { Card, CardType } from "./card";
import { Column, ColumnType, Columns, ColumnsType } from "./columns";
import { Dialog, DialogType } from "./dialog";
import { Form, FormType } from "./form";
import { Headline, HeadlineType } from "./headline";
import { Loop, LoopType } from "./loop";
import { QueryProvider, QueryProviderType } from "./queryProvider";
import { Table, TableType } from "./table";
import { Text, TextType } from "./text";

import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the type without updating the schema.
 * The schema needs to get the type input to avoid circular dependencies.
 */

const SimpleBlockUnion = z.discriminatedUnion("type", [
  Button,
  Headline,
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

export type SimpleBlocksType = ButtonType | HeadlineType | TextType | TableType;

export type ParentBlocksType =
  | CardType
  | DialogType
  | FormType
  | LoopType
  | QueryProviderType
  | ColumnType
  | ColumnsType;

export type AllBlocksType = SimpleBlocksType | ParentBlocksType;

export const AllBlocks: z.ZodType<AllBlocksType> = z.lazy(() =>
  z.union([SimpleBlockUnion, ParentBlockUnion])
);

export const TriggerBlocks = z.discriminatedUnion("type", [Button]);

export type TriggerBlocksType = z.infer<typeof TriggerBlocks>;

/* Inline blocks do not need a dialog for creation */
export const inlineBlockTypes: Set<AllBlocksType["type"]> = new Set([
  "columns",
  "card",
]);
