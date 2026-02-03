import Button from "~/design/Button";
import { UnsavedChangesHint } from "~/components/NavigationBlocker";
import { useTranslation } from "react-i18next";

type BlockEditorFooterProps = {
  formId: string;
  onCancel: () => void;
  hasChanges: boolean;
};

export const Footer = ({
  formId,
  onCancel,
  hasChanges,
}: BlockEditorFooterProps) => {
  const { t } = useTranslation();

  return (
    <div className="mt-2 flex flex-row items-center justify-between gap-2">
      <div className="px-2">{hasChanges && <UnsavedChangesHint />}</div>
      <div className="flex flex-row items-center gap-2">
        <Button variant="ghost" onClick={onCancel} type="reset">
          {t("direktivPage.blockEditor.generic.cancelButton")}
        </Button>

        <Button
          variant="primary"
          type="submit"
          form={formId}
          disabled={!hasChanges}
        >
          {t("direktivPage.blockEditor.generic.saveButton")}
        </Button>
      </div>
    </div>
  );
};
