import Button from "~/design/Button";
import { useTranslation } from "react-i18next";

type BlockEditorFooterProps = {
  onSubmit: () => void;
  onCancel: () => void;
};

export const DialogFooter = ({
  onSubmit,
  onCancel,
}: BlockEditorFooterProps) => {
  const { t } = useTranslation();

  return (
    <div>
      <Button variant="ghost" onClick={onCancel}>
        {t("direktivPage.blockEditor.generic.cancelButton")}
      </Button>
      <Button variant="primary" onClick={onSubmit}>
        {t("direktivPage.blockEditor.generic.saveButton")}
      </Button>
    </div>
  );
};
