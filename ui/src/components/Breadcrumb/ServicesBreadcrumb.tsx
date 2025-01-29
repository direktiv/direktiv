import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Layers } from "lucide-react";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const ServicesBreadcrumb = () => {
  const namespace = useNamespace();
  // const { isServicePage, isServiceDetailPage, service } =
  //   pages.services.useParams();
  const isServicePage = useMatch({
    from: "/n/$namespace/services",
    shouldThrow: false,
  });
  const { t } = useTranslation();

  if (!isServicePage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-services">
        <Link to="/n/$namespace/services" params={{ namespace }}>
          <Layers aria-hidden="true" /> {t("components.breadcrumb.services")}
        </Link>
      </BreadcrumbLink>
      {/* {isServiceDetailPage && service ? (
        <BreadcrumbLink>
          <Diamond aria-hidden="true" />
          <Link
            to={pages.services.createHref({
              namespace,
              service,
            })}
          >
            {service}
          </Link>
        </BreadcrumbLink>
      ) : null} */}
    </>
  );
};

export default ServicesBreadcrumb;
