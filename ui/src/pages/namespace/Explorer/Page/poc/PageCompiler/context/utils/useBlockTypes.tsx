import { Columns2, Heading1, LucideIcon, Text } from "lucide-react";

import { AllBlocksType } from "../../../schema/blocks";
import { BlockPathType } from "../../Block";
import { parseAncestors } from ".";
import { usePage } from "../pageCompilerContext";
import { useTranslation } from "react-i18next";

type BlockTypesConfig = {
  type: AllBlocksType["type"];
  label: string;
  icon: LucideIcon;
  allow: (path: BlockPathType) => boolean;
}[];

export const useBlockTypes = (): BlockTypesConfig => {
  const { t } = useTranslation();
  const page = usePage();

  return [
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
      allow: (path: BlockPathType) =>
        !parseAncestors({
          page,
          path,
          fn: (block) => block.type === "columns",
        }),
    },
  ];
};
