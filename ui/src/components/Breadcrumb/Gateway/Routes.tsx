import { Link, useMatch, useParams } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { SquareGanttChartIcon } from "lucide-react";

const RoutesBreadcrumb = () => {
  const { _splat } = useParams({ strict: false });
  const isGatewayRoutesPage = useMatch({
    from: "/n/$namespace/gateway/routes/",
    shouldThrow: false,
  });

  const isGatewayRoutesDetailPage = useMatch({
    from: "/n/$namespace/gateway/routes/$",
    shouldThrow: false,
  });

  if (!isGatewayRoutesPage && !isGatewayRoutesDetailPage) return null;

  return (
    <>
      {isGatewayRoutesDetailPage && (
        <BreadcrumbLink data-testid="breadcrumb-routes">
          <Link
            to="/n/$namespace/gateway/routes/$"
            from="/n/$namespace"
            params={{ _splat }}
          >
            <SquareGanttChartIcon aria-hidden="true" />
            {_splat}
          </Link>
        </BreadcrumbLink>
      )}
    </>
  );
};

export default RoutesBreadcrumb;
