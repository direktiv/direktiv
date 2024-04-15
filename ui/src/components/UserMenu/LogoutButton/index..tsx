import EnterpriseLogoutButton from "./EnterpriseLogout";
import OpenSourceLogoutButton from "./OpenSourceLogout";
import { isEnterprise } from "~/config/env/utils";

const LogoutButton = () =>
  isEnterprise ? <EnterpriseLogoutButton /> : <OpenSourceLogoutButton />;

export default LogoutButton;
