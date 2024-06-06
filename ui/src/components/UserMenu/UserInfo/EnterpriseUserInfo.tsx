import { useAuth } from "react-oidc-context";
import { useTranslation } from "react-i18next";

const EnterpriseUserInfo = () => {
  const auth = useAuth();
  const { t } = useTranslation();
  const username = auth?.user?.profile?.preferred_username ?? "";
  return <>{t("components.userMenu.loggedInAs", { username })}</>;
};

export default EnterpriseUserInfo;
