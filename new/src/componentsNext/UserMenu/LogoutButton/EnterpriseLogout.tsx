import { DropdownMenuItem } from "~/design/Dropdown";
import { LogOut } from "lucide-react";
import { useAuth } from "react-oidc-context";
import { useTranslation } from "react-i18next";

const EnterpriseLogoutButton = () => {
  const { t } = useTranslation();
  const auth = useAuth();

  const signOut = () => {
    auth.signoutRedirect();
  };

  return (
    <DropdownMenuItem className="cursor-pointer" asChild>
      <a role="button" onClick={signOut} className="flex items-center">
        <LogOut className="mr-2 h-4 w-4" />
        <span>{t("components.userMenu.logout")}</span>
      </a>
    </DropdownMenuItem>
  );
};

export default EnterpriseLogoutButton;
