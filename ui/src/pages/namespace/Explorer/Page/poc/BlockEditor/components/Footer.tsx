import Button from "~/design/Button";
import { useTranslation } from "react-i18next";

type BlockEditorFooterProps = {
  onSubmit: () => void;
  onCancel: () => void;
};

export const Footer = ({ onSubmit, onCancel }: BlockEditorFooterProps) => {
  const { t } = useTranslation();

  return (
    <div className="mt-2 flex flex-row justify-end gap-2">
      <Button variant="ghost" onClick={onCancel}>
        {t("direktivPage.blockEditor.generic.cancelButton")}
      </Button>
      <Button variant="primary" onClick={onSubmit}>
        {t("direktivPage.blockEditor.generic.saveButton")}
      </Button>
    </div>
  );
};
