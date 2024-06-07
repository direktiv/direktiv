import { useTranslation } from "react-i18next";

const OpenSourceUserInfo = () => {
  const { t } = useTranslation();
  return <>{t("components.userMenu.loggedIn")}</>;
};

export default OpenSourceUserInfo;
