import { DropdownMenuItem } from "~/design/Dropdown";
import { LogOut } from "lucide-react";
import { useApiActions } from "~/util/store/apiKey";
import { useTranslation } from "react-i18next";

const OpenSourceLogoutButton = () => {
  const { t } = useTranslation();
  const { setApiKey: storeApiKey } = useApiActions();

  const logout = () => {
    storeApiKey(null);
  };

  return (
    <DropdownMenuItem onClick={logout} className="cursor-pointer">
      <LogOut className="mr-2 h-4 w-4" />
      <span>{t("components.userMenu.logout")}</span>
    </DropdownMenuItem>
  );
};

export default OpenSourceLogoutButton;
