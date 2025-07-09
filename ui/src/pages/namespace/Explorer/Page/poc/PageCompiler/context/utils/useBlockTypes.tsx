import {
  AllBlocksType,
  FormBlocks,
  InlineBlocksType,
} from "../../../schema/blocks";
import {
  Captions,
  Columns2,
  Database,
  Heading1,
  Image,
  LucideIcon,
  RectangleHorizontal,
  Repeat2,
  Table,
  Text,
  TextCursorInput,
} from "lucide-react";

import { BlockEditFormProps } from "../../../BlockEditor";
import { BlockPathType } from "../../Block";
import { Dialog as DialogForm } from "../../../BlockEditor/Dialog";
import { DirektivPagesType } from "../../../schema";
import { Form as FormForm } from "../../../BlockEditor/Form";
import { Headline } from "../../../BlockEditor/Headline";
import { Image as ImageForm } from "../../../BlockEditor/Image";
import { Loop as LoopForm } from "../../../BlockEditor/Loop";
import { QueryProvider as QueryProviderForm } from "../../../BlockEditor/QueryProvider";
import { Table as TableForm } from "../../../BlockEditor/Table";
import { Text as TextForm } from "../../../BlockEditor/Text";
import { findAncestor } from ".";
import { useTranslation } from "react-i18next";

// common types for block type configuration
type BlockTypeConfigBase = {
  label: string;
  icon: LucideIcon;
  allow: (page: DirektivPagesType, path: BlockPathType) => boolean;
};

// inline blocks don't have a form component
type BlockTypeConfigWithoutForm = {
  [K in InlineBlocksType]: {
    type: K;
    formComponent?: never;
    defaultValues: Extract<AllBlocksType, { type: K }>;
  };
}[InlineBlocksType];

// blocks that require a form, must have a form component that implments a form for that very block type
type BlockTypeConfigWithForm = {
  [K in FormBlocks as K["type"]]: {
    type: K["type"];
    formComponent: React.ComponentType<BlockEditFormProps<K>>;
    defaultValues: K;
  };
}[FormBlocks["type"]];

type BlockTypeConfig = BlockTypeConfigBase &
  (BlockTypeConfigWithForm | BlockTypeConfigWithoutForm);

export const useBlockTypes = (): BlockTypeConfig[] => {
  const { t } = useTranslation();
  return [
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
      icon: TextCursorInput,
      allow: () => true,
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
  ];
};
