import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { Workflow } from "lucide-react";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const RoutesBreadcrumb = () => {
  const namespace = useNamespace();
  const { isGatewayRoutesPage } = pages.gateway.useParams();

  const { t } = useTranslation();

  if (!namespace) return null;
  if (!isGatewayRoutesPage) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-routes">
        <Link
          to={pages.gateway.createHref({
            namespace,
          })}
        >
          <Workflow aria-hidden="true" />
          {t("components.breadcrumb.gatewayRoutes")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default RoutesBreadcrumb;
