import {
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

import { BlockPathType } from "../../Block";
import { BlockTypeConfig } from "./types";
import { Checkbox } from "../../../BlockEditor/Form/Checkbox";
import { Dialog as DialogForm } from "../../../BlockEditor/Dialog";
import { Form as FormForm } from "../../../BlockEditor/Form";
import { Headline } from "../../../BlockEditor/Headline";
import { Image as ImageForm } from "../../../BlockEditor/Image";
import { Loop as LoopForm } from "../../../BlockEditor/Loop";
import { NumberInput } from "../../../BlockEditor/Form/NumberInput";
import { QueryProvider as QueryProviderForm } from "../../../BlockEditor/QueryProvider";
import { Select } from "../../../BlockEditor/Form/Select";
import { StringInput } from "../../../BlockEditor/Form/StringInput";
import { Table as TableForm } from "../../../BlockEditor/Table";
import { Text as TextForm } from "../../../BlockEditor/Text";
import { Textarea } from "../../../BlockEditor/Form/Textarea";
import { findAncestor } from ".";
import { t } from "i18next";
import { usePage } from "../pageCompilerContext";

const blockTypes: BlockTypeConfig[] = [
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
      },
      actions: [],
      columns: [],
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
        id: "",
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
      values: [],
      defaultValue: "",
      description: "",
      label: "",
      optional: false,
      type: "form-select",
    },
  },
];

export const useBlockTypes = () => {
  const page = usePage();

  const getBlockConfig = <T extends BlockTypeConfig["type"]>(type: T) =>
    blockTypes.find(
      (config): config is Extract<BlockTypeConfig, { type: T }> =>
        config.type === type
    );

  const getAllowedTypes = (path: BlockPathType) =>
    blockTypes.filter((type) => type.allow(page, path));

  return {
    blockTypes,
    getAllowedTypes,
    getBlockConfig,
  };
};
