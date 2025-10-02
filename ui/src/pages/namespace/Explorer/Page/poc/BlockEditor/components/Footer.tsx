import Button from "~/design/Button";
import { useTranslation } from "react-i18next";

type BlockEditorFooterProps = {
  formId: string;
  onCancel: () => void;
};

export const Footer = ({ formId, onCancel }: BlockEditorFooterProps) => {
  const { t } = useTranslation();

  return (
    <div className="mt-2 flex flex-row justify-end gap-2">
      <Button variant="ghost" onClick={onCancel} type="reset">
        {t("direktivPage.blockEditor.generic.cancelButton")}
      </Button>

      <Button variant="primary" type="submit" form={formId}>
        {t("direktivPage.blockEditor.generic.saveButton")}
      </Button>
    </div>
  );
};
