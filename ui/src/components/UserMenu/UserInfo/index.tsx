import { DropdownMenuLabel } from "~/design/Dropdown";
import EnterpriseUserInfo from "./EnterpriseUserInfo";
import OpenSourceUserInfo from "./OpenSourceUserInfo";
import { isEnterprise } from "~/config/env/utils";

const UserInfo = () => (
  <DropdownMenuLabel>
    {isEnterprise() ? <EnterpriseUserInfo /> : <OpenSourceUserInfo />}
  </DropdownMenuLabel>
);

export default UserInfo;
