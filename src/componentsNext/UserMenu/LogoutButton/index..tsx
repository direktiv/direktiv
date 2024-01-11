import EnterpriseLogoutButton from "./EnterpriseLogout";
import OpenSourceLogoutButton from "./OpenSourceLogout";

const LogoutButton = () => {
  const isEnterprise = !!process.env.VITE?.VITE_IS_ENTERPRISE;
  return isEnterprise ? <EnterpriseLogoutButton /> : <OpenSourceLogoutButton />;
};

export default LogoutButton;
