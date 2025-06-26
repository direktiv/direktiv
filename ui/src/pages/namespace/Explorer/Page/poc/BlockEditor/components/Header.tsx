import { AllBlocksType } from "../../schema/blocks";
import { BlockContextMenu } from "./ContextMenu";
import { BlockEditorAction } from "..";
import { BlockPathType } from "../../PageCompiler/Block";
import { DirektivPagesType } from "../../schema";
import { usePageEditorPanel } from "../EditorPanelProvider";
import { useTranslation } from "react-i18next";

type BlockEditDialogHeaderProps = {
  path: BlockPathType;
  action: BlockEditorAction;
  type: AllBlocksType["type"] | DirektivPagesType["type"];
};

export const DialogHeader = ({
  path,
  action,
  type,
}: BlockEditDialogHeaderProps) => {
  const { setPanel } = usePageEditorPanel();
  const { t } = useTranslation();

  return (
    <div className="flex flex-row justify-between">
      {t("direktivPage.blockEditor.editDialog.title", {
        path: path.join("."),
        action: t(`direktivPage.blockEditor.editDialog.action.${action}`),
        type: t(`direktivPage.blockEditor.editDialog.type.${type}`),
      })}
      <BlockContextMenu
        path={path}
        onDelete={() => setPanel({ action: "delete", path })}
      />
    </div>
  );
};
