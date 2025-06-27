import { DialogClose, DialogFooter } from "~/design/Dialog";

import Button from "~/design/Button";
import { DialogHeader } from "../components/Header";
import { useBlockDialog } from "../BlockDialogProvider";
import { usePageEditor } from "../../PageCompiler/context/pageCompilerContext";
import { useTranslation } from "react-i18next";

export const BlockDeleteForm = () => {
  const { t } = useTranslation();
  const { dialog } = useBlockDialog();
  const { deleteBlock } = usePageEditor();

  if (dialog?.action !== "delete")
    throw new Error("BlockDeleteForm used with invalid dialog state");

  return (
    <>
      <DialogHeader />
      <div>{t("direktivPage.blockEditor.delete.warning")}</div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("direktivPage.blockEditor.generic.cancelButton")}
          </Button>
        </DialogClose>
        <DialogClose asChild>
          <Button variant="primary" onClick={() => deleteBlock(dialog.path)}>
            {t("direktivPage.blockEditor.generic.confirmButton")}
          </Button>
        </DialogClose>
      </DialogFooter>
    </>
  );
};
