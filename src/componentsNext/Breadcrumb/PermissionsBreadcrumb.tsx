import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const PermissionsBreadcrumb = () => {
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
    </>
  );
};

export default PermissionsBreadcrumb;
