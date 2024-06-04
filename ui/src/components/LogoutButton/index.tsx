import EnterpriseLogoutButton from "./EnterpriseLogout";
import { FC } from "react";
import { LogoutButtonProps } from "./types";
import OpenSourceLogoutButton from "./OpenSourceLogout";
import { isEnterprise } from "~/config/env/utils";

const LogoutButton: FC<LogoutButtonProps> = ({ children, button }) =>
  isEnterprise() ? (
    <EnterpriseLogoutButton button={button}>{children}</EnterpriseLogoutButton>
  ) : (
    <OpenSourceLogoutButton button={button}>{children}</OpenSourceLogoutButton>
  );

export default LogoutButton;
