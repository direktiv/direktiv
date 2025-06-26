import {
  DialogFooter as DesignDialogFooter,
  DialogClose,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { useTranslation } from "react-i18next";

type BlockEditorFooterProps = {
  formId: string;
};

export const DialogFooter = ({ formId }: BlockEditorFooterProps) => {
  const { t } = useTranslation();

  return (
    <DesignDialogFooter>
      <DialogClose asChild>
        <Button variant="ghost">
          {t("direktivPage.blockEditor.generic.cancelButton")}
        </Button>
      </DialogClose>
      <DialogClose asChild>
        <Button variant="primary" type="submit" form={formId}>
          {t("direktivPage.blockEditor.generic.saveButton")}
        </Button>
      </DialogClose>
    </DesignDialogFooter>
  );
};
