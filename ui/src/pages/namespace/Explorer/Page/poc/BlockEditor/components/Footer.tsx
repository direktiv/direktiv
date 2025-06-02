import {
  DialogFooter as DesignDialogFooter,
  DialogClose,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { useTranslation } from "react-i18next";

type BlockEditorFooterProps = {
  onSave: () => void;
};

export const DialogFooter = ({ onSave }: BlockEditorFooterProps) => {
  const { t } = useTranslation();

  return (
    <DesignDialogFooter>
      <DialogClose>
        <Button variant="ghost">Cancel</Button>
      </DialogClose>
      <DialogClose>
        <Button variant="primary" onClick={onSave}>
          {t("direktivPage.blockEditor.generic.saveButton")}
        </Button>
      </DialogClose>
    </DesignDialogFooter>
  );
};
