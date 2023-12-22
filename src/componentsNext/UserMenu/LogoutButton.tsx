import { DropdownMenuItem } from "~/design/Dropdown";
import { LogOut } from "lucide-react";
import enterpriseConfig from "~/config/enterprise";
import env from "~/config/env";
import { useApiActions } from "~/util/store/apiKey";
import { useTranslation } from "react-i18next";

const LogoutButton = () => {
  const { t } = useTranslation();
  const { setApiKey: storeApiKey } = useApiActions();

  const { VITE_IS_ENTERPRISE: isEnterprise } = env;

  const logout = () => {
    storeApiKey(null);
  };

  return isEnterprise === "true" ? (
    <DropdownMenuItem className="cursor-pointer" asChild>
      <a href={enterpriseConfig.logoutPath} className="flex items-center">
        <LogOut className="mr-2 h-4 w-4" />
        <span>{t("components.userMenu.logout")}</span>
      </a>
    </DropdownMenuItem>
  ) : (
    <DropdownMenuItem onClick={logout} className="cursor-pointer">
      <LogOut className="mr-2 h-4 w-4" />
      <span>{t("components.userMenu.logout")}</span>
    </DropdownMenuItem>
  );
};

export default LogoutButton;
