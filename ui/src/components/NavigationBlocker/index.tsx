import { FC } from "react";
import useNavigationBlocker from "~/hooks/useNavigationBlocker";
import { useTranslation } from "react-i18next";

export const NavigationBlocker: FC = () => {
  const { t } = useTranslation();
  useNavigationBlocker(t("components.blocker.unsavedChangesWarning"));

  return null;
};

export const UnsavedChangesHint = () => {
  const { t } = useTranslation();

  return (
    <div className="text-sm text-gray-8 dark:text-gray-dark-8">
      <span className="text-center">
        {t("pages.explorer.workflow.editor.unsavedNote")}
      </span>
    </div>
  );
};
