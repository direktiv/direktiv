import { useTranslation } from "react-i18next";

export const UnsavedChanges = () => {
  const { t } = useTranslation();

  return (
    <div className="text-sm text-gray-8 dark:text-gray-dark-8">
      <span className="text-center">
        {t("pages.explorer.workflow.editor.unsavedNote")}
      </span>
    </div>
  );
};
