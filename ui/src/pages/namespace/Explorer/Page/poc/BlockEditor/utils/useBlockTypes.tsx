import {
  Calendar,
  Captions,
  ChevronsUpDown,
  Columns2,
  Database,
  FileText,
  Heading1,
  Image,
  RectangleHorizontal,
  Repeat2,
  SquareCheck,
  Table,
  Text,
  TextCursorInput,
  Type,
} from "lucide-react";

import { BlockPathType } from "../../PageCompiler/Block";
import { BlockTypeConfig } from "./types";
import { Checkbox } from "../Form/Checkbox";
import { DateInput } from "../Form/DateInput";
import { Dialog as DialogForm } from "../Dialog";
import { Form as FormForm } from "../Form";
import { Headline } from "../Headline";
import { Image as ImageForm } from "../Image";
import { Loop as LoopForm } from "../Loop";
import { NumberInput } from "../Form/NumberInput";
import { QueryProvider as QueryProviderForm } from "../QueryProvider";
import { Select } from "../Form/Select";
import { StringInput } from "../Form/StringInput";
import { Table as TableForm } from "../Table";
import { Text as TextForm } from "../Text";
import { Textarea } from "../Form/Textarea";
import { findAncestor } from ".";
import { t } from "i18next";
import { useCallback } from "react";
import { usePage } from "../../PageCompiler/context/pageCompilerContext";

export const blockTypes: BlockTypeConfig[] = [
  {
    type: "headline",
    label: t("direktivPage.blockEditor.blockName.headline"),
    icon: Heading1,
    allow: () => true,
    formComponent: Headline,
    defaultValues: {
      type: "headline",
      level: "h1",
      label: "",
    },
  },
  {
    type: "text",
    label: t("direktivPage.blockEditor.blockName.text"),
    icon: Text,
    allow: () => true,
    formComponent: TextForm,
    defaultValues: { type: "text", content: "" },
  },
  {
    type: "query-provider",
    label: t("direktivPage.blockEditor.blockName.query-provider"),
    icon: Database,
    allow: () => true,
    formComponent: QueryProviderForm,
    defaultValues: { type: "query-provider", blocks: [], queries: [] },
  },
  {
    type: "columns",
    label: t("direktivPage.blockEditor.blockName.columns"),
    icon: Columns2,
    allow: (page, path) =>
      !findAncestor({
        page,
        path,
        match: (block) => block.type === "columns",
      }),
    defaultValues: {
      type: "columns",
      blocks: [
        {
          type: "column",
          blocks: [],
        },
        {
          type: "column",
          blocks: [],
        },
      ],
    },
  },
  {
    type: "card",
    label: t("direktivPage.blockEditor.blockName.card"),
    icon: Captions,
    allow: (page, path) =>
      !findAncestor({
        page,
        path,
        match: (block) => block.type === "card",
      }),
    defaultValues: { type: "card", blocks: [] },
  },
  {
    type: "table",
    label: t("direktivPage.blockEditor.blockName.table"),
    icon: Table,
    allow: () => true,
    formComponent: TableForm,
    defaultValues: {
      type: "table",
      data: {
        type: "loop",
        id: "",
        data: "",
        pageSize: 10,
      },
      columns: [],
      blocks: [
        {
          type: "table-actions",
          blocks: [],
        },
        {
          type: "row-actions",
          blocks: [],
        },
      ],
    },
  },
  {
    type: "dialog",
    label: t("direktivPage.blockEditor.blockName.dialog"),
    icon: RectangleHorizontal,
    allow: (page, path) =>
      !findAncestor({
        page,
        path,
        match: (block) => block.type === "dialog",
      }),
    formComponent: DialogForm,
    defaultValues: {
      type: "dialog",
      trigger: {
        type: "button",
        label: "",
      },
      blocks: [],
    },
  },
  {
    type: "loop",
    label: t("direktivPage.blockEditor.blockName.loop"),
    icon: Repeat2,
    allow: () => true,
    formComponent: LoopForm,
    defaultValues: {
      type: "loop",
      id: "",
      data: "",
      blocks: [],
      pageSize: 10,
    },
  },
  {
    type: "image",
    label: t("direktivPage.blockEditor.blockName.image"),
    icon: Image,
    allow: () => true,
    formComponent: ImageForm,
    defaultValues: { type: "image", src: "", width: 200, height: 200 },
  },
  {
    type: "form",
    label: t("direktivPage.blockEditor.blockName.form"),
    icon: FileText,
    allow: (page, path) =>
      !findAncestor({
        page,
        path,
        match: (block) => block.type === "form",
      }),
    formComponent: FormForm,
    defaultValues: {
      type: "form",
      mutation: {
        url: "",
        method: "POST",
      },
      trigger: {
        label: "",
        type: "button",
      },
      blocks: [],
    },
  },
  {
    type: "form-string-input",
    label: t("direktivPage.blockEditor.blockName.form-string-input"),
    icon: TextCursorInput,
    allow: (page, path) =>
      !!findAncestor({
        page,
        path,
        match: (block) => block.type === "form",
      }),
    formComponent: StringInput,
    defaultValues: {
      id: "",
      defaultValue: "",
      description: "",
      label: "",
      optional: false,
      type: "form-string-input",
      variant: "text",
    },
  },
  {
    type: "form-number-input",
    label: t("direktivPage.blockEditor.blockName.form-number-input"),
    icon: TextCursorInput,
    allow: (page, path) =>
      !!findAncestor({
        page,
        path,
        match: (block) => block.type === "form",
      }),
    formComponent: NumberInput,
    defaultValues: {
      id: "",
      defaultValue: { type: "number", value: 0 },
      description: "",
      label: "",
      optional: false,
      type: "form-number-input",
    },
  },
  {
    type: "form-date-input",
    label: t("direktivPage.blockEditor.blockName.form-date-input"),
    icon: Calendar,
    allow: (page, path) =>
      !!findAncestor({
        page,
        path,
        match: (block) => block.type === "form",
      }),
    formComponent: DateInput,
    defaultValues: {
      id: "",
      defaultValue: "",
      description: "",
      label: "",
      optional: false,
      type: "form-date-input",
    },
  },
  {
    type: "form-checkbox",
    label: t("direktivPage.blockEditor.blockName.form-checkbox"),
    icon: SquareCheck,
    allow: (page, path) =>
      !!findAncestor({
        page,
        path,
        match: (block) => block.type === "form",
      }),
    formComponent: Checkbox,
    defaultValues: {
      id: "",
      defaultValue: { type: "boolean", value: false },
      description: "",
      label: "",
      optional: false,
      type: "form-checkbox",
    },
  },
  {
    type: "form-textarea",
    label: t("direktivPage.blockEditor.blockName.form-textarea"),
    icon: Type,
    allow: (page, path) =>
      !!findAncestor({
        page,
        path,
        match: (block) => block.type === "form",
      }),
    formComponent: Textarea,
    defaultValues: {
      id: "",
      defaultValue: "",
      description: "",
      label: "",
      optional: false,
      type: "form-textarea",
    },
  },
  {
    type: "form-select",
    label: t("direktivPage.blockEditor.blockName.form-select"),
    icon: ChevronsUpDown,
    allow: (page, path) =>
      !!findAncestor({
        page,
        path,
        match: (block) => block.type === "form",
      }),
    formComponent: Select,
    defaultValues: {
      id: "",
      values: { type: "static-select-options", value: [] },
      defaultValue: "",
      description: "",
      label: "",
      optional: false,
      type: "form-select",
    },
  },
];

export const getBlockConfig = <T extends BlockTypeConfig["type"]>(type: T) =>
  blockTypes.find(
    (config): config is Extract<BlockTypeConfig, { type: T }> =>
      config.type === type
  );

export const useAllowedBlockTypes = () => {
  const page = usePage();

  return useCallback(
    (path: BlockPathType) =>
      blockTypes.filter((type) => type.allow(page, path)),
    [page]
  );
};
