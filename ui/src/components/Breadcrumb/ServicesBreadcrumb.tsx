import { Diamond, Layers } from "lucide-react";
import { Link, useMatch, useParams } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { useTranslation } from "react-i18next";

const ServicesBreadcrumb = () => {
  const { service } = useParams({ strict: false });

  const isServicePage = useMatch({
    from: "/n/$namespace/services/",
    shouldThrow: false,
  });

  const isServiceDetailPage = useMatch({
    from: "/n/$namespace/services/$service",
    shouldThrow: false,
  });

  const isServicesRoute = isServicePage || isServiceDetailPage;

  const { t } = useTranslation();

  if (!isServicesRoute) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-services">
        <Link to="/n/$namespace/services" from="/n/$namespace">
          <Layers aria-hidden="true" /> {t("components.breadcrumb.services")}
        </Link>
      </BreadcrumbLink>
      {isServiceDetailPage && service ? (
        <BreadcrumbLink>
          <Diamond aria-hidden="true" />
          <Link
            to="/n/$namespace/services/$service"
            from="/n/$namespace"
            params={{ service }}
          >
            {service}
          </Link>
        </BreadcrumbLink>
      ) : null}
    </>
  );
};

export default ServicesBreadcrumb;
