import {
  Captions,
  Columns2,
  Database,
  Heading1,
  LucideIcon,
  Table,
  Text,
} from "lucide-react";

import { AllBlocksType } from "../../../schema/blocks";
import { BlockPathType } from "../../Block";
import { findInBranch } from ".";
import { usePage } from "../pageCompilerContext";
import { useTranslation } from "react-i18next";

type BlockTypeConfig = {
  type: AllBlocksType["type"];
  label: string;
  icon: LucideIcon;
  allow: boolean;
};

type BlockTypeConfigReturn = Omit<BlockTypeConfig, "allow">;

export const useBlockTypes = (path: BlockPathType): BlockTypeConfigReturn[] => {
  const { t } = useTranslation();
  const page = usePage();

  const config = [
    {
      type: "headline",
      label: t("direktivPage.blockEditor.blockName.headline"),
      icon: Heading1,
      allow: true,
    },
    {
      type: "text",
      label: t("direktivPage.blockEditor.blockName.text"),
      icon: Text,
      allow: true,
    },
    {
      type: "query-provider",
      label: t("direktivPage.blockEditor.blockName.query-provider"),
      icon: Database,
      allow: true,
    },
    {
      type: "columns",
      label: t("direktivPage.blockEditor.blockName.columns"),
      icon: Columns2,
      allow: !findInBranch({
        page,
        path,
        match: (block) => block.type === "columns",
      }),
    },
    {
      type: "card",
      label: t("direktivPage.blockEditor.blockName.card"),
      icon: Captions,
      allow: !findInBranch({
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
    },
  ] satisfies BlockTypeConfig[];

  return config
    .filter((type) => type.allow)
    .map(({ allow: _, ...rest }) => rest);
};
