import {
  DialogFooter as DesignDialogFooter,
  DialogClose,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { useTranslation } from "react-i18next";

type BlockEditorFooterProps = {
  onSubmit: () => void;
};

export const DialogFooter = ({ onSubmit }: BlockEditorFooterProps) => {
  const { t } = useTranslation();

  return (
    <DesignDialogFooter>
      <DialogClose asChild>
        <Button variant="ghost">
          {t("direktivPage.blockEditor.generic.cancelButton")}
        </Button>
      </DialogClose>
      <DialogClose asChild>
        <Button variant="primary" onClick={onSubmit}>
          {t("direktivPage.blockEditor.generic.saveButton")}
        </Button>
      </DialogClose>
    </DesignDialogFooter>
  );
};
