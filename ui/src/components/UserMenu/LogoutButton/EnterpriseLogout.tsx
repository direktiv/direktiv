import { DropdownMenuItem } from "~/design/Dropdown";
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
    <DropdownMenuItem onClick={logout} className="cursor-pointer">
      <LogOut className="mr-2 h-4 w-4" />
      <span>{t("components.userMenu.logout")}</span>
    </DropdownMenuItem>
  );
};

export default EnterpriseLogoutButton;
