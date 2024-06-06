import {
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
} from "~/design/Dropdown";

import EnterpriseUserInfo from "./EnterpriseUserInfo";
import { LogOut } from "lucide-react";
import LogoutButton from "~/components/LogoutButton";
import OpenSourceUserInfo from "./OpenSourceUserInfo";
import { isEnterprise } from "~/config/env/utils";
import useApiKeyHandling from "~/hooks/useApiKeyHandling";
import { useTranslation } from "react-i18next";

const UserInfo = () => {
  const { usesAccounts } = useApiKeyHandling();
  const { t } = useTranslation();
  return usesAccounts ? (
    <>
      <DropdownMenuLabel>
        {isEnterprise() ? <EnterpriseUserInfo /> : <OpenSourceUserInfo />}
      </DropdownMenuLabel>
      <DropdownMenuSeparator />
      <LogoutButton
        button={(props) => (
          <DropdownMenuItem {...props} className="cursor-pointer">
            <LogOut className="mr-2 h-4 w-4" />
            <span>{t("components.userMenu.logout")}</span>
          </DropdownMenuItem>
        )}
      />
      <DropdownMenuSeparator />
    </>
  ) : null;
};

export default UserInfo;
