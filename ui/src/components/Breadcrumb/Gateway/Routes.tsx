import { SquareGanttIcon, Workflow } from "lucide-react";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const RoutesBreadcrumb = () => {
  const pages = usePages();
  const namespace = useNamespace();
  const { isGatewayRoutesPage, isGatewayRoutesDetailPage, routePath } =
    pages.gateway.useParams();

  const { t } = useTranslation();

  if (!namespace) return null;
  if (!isGatewayRoutesPage && !isGatewayRoutesDetailPage) return null;

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
      {routePath && (
        <BreadcrumbLink data-testid="breadcrumb-routes">
          <Link
            to={pages.gateway.createHref({
              namespace,
              subpage: "routeDetail",
              routePath,
            })}
          >
            <SquareGanttIcon aria-hidden="true" />
            {routePath}
          </Link>
        </BreadcrumbLink>
      )}
    </>
  );
};

export default RoutesBreadcrumb;
