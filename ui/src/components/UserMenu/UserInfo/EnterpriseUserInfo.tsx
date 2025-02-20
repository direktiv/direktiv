import { useAuth } from "react-oidc-context";
import { useTranslation } from "react-i18next";

const EnterpriseUserInfo = () => {
  const auth = useAuth();
  const { t } = useTranslation();

  const username =
    auth?.user?.profile?.preferred_username ??
    auth?.user?.profile?.name ??
    auth?.user?.profile?.email;

  if (!username) {
    return <>{t("components.userMenu.loggedIn")}</>;
  }

  return <>{t("components.userMenu.loggedInAs", { username })}</>;
};

export default EnterpriseUserInfo;
