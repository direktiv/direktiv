import {
  DialogHeader as DesignDialogHeader,
  DialogTitle,
} from "~/design/Dialog";

import { AllBlocksType } from "../../schema/blocks";
import { BlockEditorAction } from "..";
import { BlockPathType } from "../../PageCompiler/Block";
import { useTranslation } from "react-i18next";

type BlockEditDialogHeaderProps = {
  path: BlockPathType;
  action: BlockEditorAction;
  type: AllBlocksType["type"];
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
        {t("direktivPage.blockEditor.dialogTitle", {
          path: path.join("."),
          action: t(`direktivPage.blockEditor.action.${action}`),
          type: t(`direktivPage.blockEditor.blockType.${type}`),
        })}
      </DialogTitle>
    </DesignDialogHeader>
  );
};
