import { Link, Outlet } from "react-router-dom";

import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";

const PermissionsPage = () => {
  const test = pages.permissions?.createHref;

  const namespace = useNamespace();

  if (!test) return null;
  if (!namespace) return null;

  return (
    <div>
      <div className="flex gap-5">
        <Link
          to={test({
            namespace,
          })}
        >
          policy
        </Link>
        <Link
          to={test({
            namespace,
            subpage: "groups",
          })}
        >
          Groups
        </Link>
        <Link
          to={test({
            namespace,
            subpage: "tokens",
          })}
        >
          Tokens
        </Link>
      </div>
      <Outlet />
    </div>
  );
};

export default PermissionsPage;
