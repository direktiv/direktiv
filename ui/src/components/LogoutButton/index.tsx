import EnterpriseLogoutButton from "./EnterpriseLogout";
import { FC } from "react";
import { LogoutButtonProps } from "./types";
import OpenSourceLogoutButton from "./OpenSourceLogout";
import { isEnterprise } from "~/config/env/utils";

const LogoutButton: FC<LogoutButtonProps> = ({ children, wrapper }) =>
  isEnterprise() ? (
    <EnterpriseLogoutButton wrapper={wrapper}>
      {children}
    </EnterpriseLogoutButton>
  ) : (
    <OpenSourceLogoutButton wrapper={wrapper}>
      {children}
    </OpenSourceLogoutButton>
  );

export default LogoutButton;
