import { BlockContextMenu } from "./ContextMenu";
import { BlockEditorAction } from "..";
import { BlockPathType } from "../../PageCompiler/Block";
import { BlockType } from "../../schema/blocks";
import { usePageEditorPanel } from "../EditorPanelProvider";
import { useTranslation } from "react-i18next";

type BlockFormHeaderProps = {
  path: BlockPathType;
  action: BlockEditorAction;
  block: BlockType;
};

export const Header = ({ path, action, block }: BlockFormHeaderProps) => {
  const { panel, setPanel } = usePageEditorPanel();
  const { t } = useTranslation();

  if (!panel) return null;

  return (
    <div className="flex flex-row justify-between text-lg font-semibold">
      {t("direktivPage.blockEditor.blockForm.title", {
        path: path.join("."),
        action: t(`direktivPage.blockEditor.blockForm.action.${action}`),
        type: t(`direktivPage.blockEditor.blockForm.type.${block.type}`),
      })}
      <BlockContextMenu
        path={path}
        onDelete={() => setPanel({ ...panel, action: "delete" })}
      />
    </div>
  );
};
