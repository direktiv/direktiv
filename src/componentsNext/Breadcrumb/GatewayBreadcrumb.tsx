import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const GatewayBreadcrumb = () => {
  const namespace = useNamespace();
  const { isGatewayPage } = pages.gateway.useParams();
  const { icon: Icon } = pages.gateway;
  const { t } = useTranslation();

  if (!isGatewayPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={pages.gateway.createHref({
            namespace,
          })}
        >
          <Icon aria-hidden="true" />
          {t("components.breadcrumb.gateway")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default GatewayBreadcrumb;
