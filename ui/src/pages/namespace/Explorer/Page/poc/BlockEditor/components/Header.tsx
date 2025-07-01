import { AllBlocksType } from "../../schema/blocks";
import { BlockContextMenu } from "./ContextMenu";
import { BlockEditorAction } from "..";
import { BlockPathType } from "../../PageCompiler/Block";
import { DirektivPagesType } from "../../schema";
import { usePageEditorPanel } from "../EditorPanelProvider";
import { useTranslation } from "react-i18next";

type BlockBlockFormHeaderProps = {
  path: BlockPathType;
  action: BlockEditorAction;
  type: AllBlocksType["type"] | DirektivPagesType["type"];
};

export const Header = ({ path, action, type }: BlockBlockFormHeaderProps) => {
  const { setPanel } = usePageEditorPanel();
  const { t } = useTranslation();

  return (
    <div className="flex flex-row justify-between">
      {t("direktivPage.blockEditor.blockForm.title", {
        path: path.join("."),
        action: t(`direktivPage.blockEditor.blockForm.action.${action}`),
        type: t(`direktivPage.blockEditor.blockForm.type.${type}`),
      })}
      <BlockContextMenu
        path={path}
        onDelete={() => setPanel({ action: "delete", path })}
      />
    </div>
  );
};
