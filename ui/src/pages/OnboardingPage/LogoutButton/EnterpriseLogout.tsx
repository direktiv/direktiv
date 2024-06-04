import Button from "~/design/Button";
import { LogOut } from "lucide-react";
import { useAuth } from "react-oidc-context";
import { useTranslation } from "react-i18next";

const EnterpriseLogoutButton = () => {
  const { t } = useTranslation();
  const auth = useAuth();

  const logout = () => {
    auth.signoutRedirect();
  };

  return (
    <Button variant="outline" onClick={logout}>
      <LogOut />
      {t("pages.onboarding.logout")}
    </Button>
  );
};

export default EnterpriseLogoutButton;
