import {
  DialogHeader as DesignDialogHeader,
  DialogTitle,
} from "~/design/Dialog";

import { AllBlocksType } from "../../schema/blocks";
import { BlockEditorAction } from "..";
import { BlockPathType } from "../../PageCompiler/Block";
import { DirektivPagesType } from "../../schema";
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
  const { t } = useTranslation();
  return (
    <DesignDialogHeader>
      <DialogTitle>
        {t("direktivPage.blockEditor.editDialog.title", {
          path: path.join("."),
          action: t(`direktivPage.blockEditor.editDialog.action.${action}`),
          type: t(`direktivPage.blockEditor.editDialog.type.${type}`),
        })}
      </DialogTitle>
    </DesignDialogHeader>
  );
};
