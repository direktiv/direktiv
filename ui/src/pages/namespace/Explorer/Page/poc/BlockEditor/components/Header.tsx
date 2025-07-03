import { AllBlocksType } from "../../schema/blocks";
import { BlockContextMenu } from "./ContextMenu";
import { BlockEditorAction } from "..";
import { BlockPathType } from "../../PageCompiler/Block";
import { usePageEditorPanel } from "../EditorPanelProvider";
import { useTranslation } from "react-i18next";

type BlockFormHeaderProps = {
  path: BlockPathType;
  action: BlockEditorAction;
  block: AllBlocksType;
};

export const Header = ({ path, action, block }: BlockFormHeaderProps) => {
  const { setPanel } = usePageEditorPanel();
  const { t } = useTranslation();

  return (
    <div className="flex flex-row justify-between text-lg font-semibold">
      {t("direktivPage.blockEditor.blockForm.title", {
        path: path.join("."),
        action: t(`direktivPage.blockEditor.blockForm.action.${action}`),
        type: t(`direktivPage.blockEditor.blockForm.type.${block.type}`),
      })}
      <BlockContextMenu
        path={path}
        onDelete={() => setPanel({ action: "delete", path, block })}
      />
    </div>
  );
};
