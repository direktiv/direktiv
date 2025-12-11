import Button from "~/design/Button";
import { useTranslation } from "react-i18next";

type BlockEditorFooterProps = {
  formId: string;
  onCancel: () => void;
  disableSave: boolean;
};

export const Footer = ({
  formId,
  onCancel,
  disableSave,
}: BlockEditorFooterProps) => {
  const { t } = useTranslation();

  return (
    <div className="mt-2 flex flex-row justify-end gap-2">
      <Button variant="ghost" onClick={onCancel} type="reset">
        {t("direktivPage.blockEditor.generic.cancelButton")}
      </Button>

      <Button
        variant="primary"
        type="submit"
        form={formId}
        disabled={disableSave}
      >
        {t("direktivPage.blockEditor.generic.saveButton")}
      </Button>
    </div>
  );
};
