import {
  DialogHeader as DesignDialogHeader,
  DialogTitle,
} from "~/design/Dialog";

import { useBlockDialog } from "../BlockDialogProvider";
import { useTranslation } from "react-i18next";

export const DialogHeader = () => {
  const { t } = useTranslation();
  const { dialog } = useBlockDialog();

  if (!dialog) return null;

  const { action, block, path } = dialog;

  return (
    <DesignDialogHeader>
      <DialogTitle>
        {t("direktivPage.blockEditor.editDialog.title", {
          path: path.join("."),
          action: t(`direktivPage.blockEditor.editDialog.action.${action}`),
          type: t(`direktivPage.blockEditor.editDialog.type.${block.type}`),
        })}
      </DialogTitle>
    </DesignDialogHeader>
  );
};
