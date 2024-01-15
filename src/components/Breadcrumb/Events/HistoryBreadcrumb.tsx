import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { Radio } from "lucide-react";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const EventsListenerBreadcrumb = () => {
  const namespace = useNamespace();
  const { isEventsHistoryPage } = pages.events.useParams();
  const { t } = useTranslation();

  if (!isEventsHistoryPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-event-history">
        <Link
          to={pages.events.createHref({
            namespace,
          })}
        >
          <Radio aria-hidden="true" />
          {t("components.breadcrumb.eventHistory")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default EventsListenerBreadcrumb;
