import { Columns2, Heading1, LucideIcon, Text } from "lucide-react";

import { AllBlocksType } from "../../../schema/blocks";
import { BlockPathType } from "../../Block";
import { parseAncestors } from ".";
import { usePage } from "../pageCompilerContext";
import { useTranslation } from "react-i18next";

type BlockTypeConfig = {
  type: AllBlocksType["type"];
  label: string;
  icon: LucideIcon;
  allow: () => boolean;
};

export const useBlockTypes = (path: BlockPathType): BlockTypeConfig[] => {
  const { t } = useTranslation();
  const page = usePage();

  const config = [
    {
      type: "headline",
      label: t("direktivPage.blockEditor.blockName.headline"),
      icon: Heading1,
      allow: () => true,
    },
    {
      type: "text",
      label: t("direktivPage.blockEditor.blockName.text"),
      icon: Text,
      allow: () => true,
    },
    {
      type: "columns",
      label: t("direktivPage.blockEditor.blockName.columns"),
      icon: Columns2,
      allow: () =>
        !parseAncestors({
          page,
          path,
          fn: (block) => block.type === "columns",
        }),
    },
  ] satisfies BlockTypeConfig[];

  return config.filter((type) => type.allow() === true);
};
