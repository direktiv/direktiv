import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const MonitoringBreadcrumb = () => {
  const pages = usePages();
  const namespace = useNamespace();
  const { isMonitoringPage } = pages.monitoring.useParams();
  const { icon: Icon } = pages.monitoring;
  const { t } = useTranslation();

  if (!isMonitoringPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={pages.monitoring.createHref({
            namespace,
          })}
        >
          <Icon aria-hidden="true" />
          {t("components.breadcrumb.monitoring")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default MonitoringBreadcrumb;
