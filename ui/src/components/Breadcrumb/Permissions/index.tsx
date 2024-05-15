import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import GroupsBreadcrumb from "./GroupsBreadcrumb";
import { Link } from "react-router-dom";
import PolicyBreadcrumb from "./PolicyBreadcumb";
import TokensBreadcrumb from "./TokenBreadCrumb";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const PermissionsBreadcrumb = () => {
  const pages = usePages();
  const namespace = useNamespace();
  const { t } = useTranslation();

  const permissions = pages.permissions;
  if (!permissions) return null;

  const { isPermissionsPage } = permissions.useParams();
  const { icon: Icon } = permissions;

  if (!isPermissionsPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={permissions.createHref({
            namespace,
          })}
        >
          <Icon aria-hidden="true" />
          {t("components.breadcrumb.permissions")}
        </Link>
      </BreadcrumbLink>
      <PolicyBreadcrumb />
      <GroupsBreadcrumb />
      <TokensBreadcrumb />
    </>
  );
};

export default PermissionsBreadcrumb;
