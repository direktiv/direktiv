import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Diamond } from "lucide-react";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const ServicesBreadcrumb = () => {
  const namespace = useNamespace();
  const { isServicePage, isServiceDetailPage, service } =
    pages.services.useParams();
  const { icon: Icon } = pages.services;
  const { t } = useTranslation();

  if (!isServicePage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={pages.services.createHref({
            namespace,
          })}
        >
          <Icon aria-hidden="true" />
          {t("components.breadcrumb.services")}
        </Link>
      </BreadcrumbLink>
      {isServiceDetailPage && service ? (
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
      ) : null}
    </>
  );
};

export default ServicesBreadcrumb;
