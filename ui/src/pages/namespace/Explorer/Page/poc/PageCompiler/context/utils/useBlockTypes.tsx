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
import { FormBlocks, InlineBlocksType } from "../../../schema/blocks";

import { BlockEditFormProps } from "../../../BlockEditor";
import { BlockPathType } from "../../Block";
import { Dialog as DialogForm } from "../../../BlockEditor/Dialog";
import { Form as FormForm } from "../../../BlockEditor/Form";
import { Headline } from "../../../BlockEditor/Headline";
import { Image as ImageForm } from "../../../BlockEditor/Image";
import { Loop as LoopForm } from "../../../BlockEditor/Loop";
import { QueryProvider as QueryProviderForm } from "../../../BlockEditor/QueryProvider";
import { Table as TableForm } from "../../../BlockEditor/Table";
import { Text as TextForm } from "../../../BlockEditor/Text";
import { findAncestor } from ".";
import { usePage } from "../pageCompilerContext";
import { useTranslation } from "react-i18next";

// common types for block type configuration
type BlockTypeConfigBase = {
  label: string;
  icon: LucideIcon;
  allow: boolean;
};

// inline blocks don't have a form component
type BlockTypeConfigWithoutForm = BlockTypeConfigBase & {
  type: InlineBlocksType;
  formComponent?: never;
};

// blocks that require a form, must have a form component that implments a form for that very block type
type BlockTypeConfigWithForm = {
  [K in FormBlocks as K["type"]]: BlockTypeConfigBase & {
    type: K["type"];
    formComponent: React.ComponentType<BlockEditFormProps<K>>;
  };
}[FormBlocks["type"]];

type BlockTypeConfig = BlockTypeConfigWithForm | BlockTypeConfigWithoutForm;

export const useBlockTypes = (path: BlockPathType) => {
  const { t } = useTranslation();
  const page = usePage();

  const config: BlockTypeConfig[] = [
    {
      type: "headline",
      label: t("direktivPage.blockEditor.blockName.headline"),
      icon: Heading1,
      allow: true,
      formComponent: Headline,
    },
    {
      type: "text",
      label: t("direktivPage.blockEditor.blockName.text"),
      icon: Text,
      allow: true,
      formComponent: TextForm,
    },
    {
      type: "query-provider",
      label: t("direktivPage.blockEditor.blockName.query-provider"),
      icon: Database,
      allow: true,
      formComponent: QueryProviderForm,
    },
    {
      type: "columns",
      label: t("direktivPage.blockEditor.blockName.columns"),
      icon: Columns2,
      allow: !findAncestor({
        page,
        path,
        match: (block) => block.type === "columns",
      }),
    },
    {
      type: "card",
      label: t("direktivPage.blockEditor.blockName.card"),
      icon: Captions,
      allow: !findAncestor({
        page,
        path,
        match: (block) => block.type === "card",
      }),
    },
    {
      type: "table",
      label: t("direktivPage.blockEditor.blockName.table"),
      icon: Table,
      allow: true,
      formComponent: TableForm,
    },
    {
      type: "dialog",
      label: t("direktivPage.blockEditor.blockName.dialog"),
      icon: RectangleHorizontal,
      allow: !findAncestor({
        page,
        path,
        match: (block) => block.type === "dialog",
      }),
      formComponent: DialogForm,
    },
    {
      type: "loop",
      label: t("direktivPage.blockEditor.blockName.loop"),
      icon: Repeat2,
      allow: true,
      formComponent: LoopForm,
    },
    {
      type: "image",
      label: t("direktivPage.blockEditor.blockName.image"),
      icon: Image,
      allow: true,
      formComponent: ImageForm,
    },
    {
      type: "form",
      label: t("direktivPage.blockEditor.blockName.form"),
      icon: TextCursorInput,
      allow: true,
      formComponent: FormForm,
    },
  ] as const;

  return config
    .filter((type) => type.allow)
    .map(({ allow: _, ...rest }) => rest);
};
