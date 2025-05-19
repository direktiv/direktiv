import { DialogHeader, DialogTitle } from "~/design/Dialog";

import { BlockEditFormProps } from ".";
import { useTranslation } from "react-i18next";

export const Text = ({ block, path }: BlockEditFormProps) => {
  const { t } = useTranslation();

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          {t("direktivPage.blockEditor.Text.modalTitle", {
            path: path.join("."),
          })}
        </DialogTitle>
      </DialogHeader>
      <div>{JSON.stringify(block)}</div>
    </>
  );
};
