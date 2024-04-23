import { FC } from "react";
import useNavigationBlocker from "~/hooks/useNavigationBlocker";
import { useTranslation } from "react-i18next";

const NavigationBlocker: FC = () => {
  const { t } = useTranslation();
  useNavigationBlocker(t("components.blocker.unsavedChangesWarning"));

  return null;
};

export default NavigationBlocker;
