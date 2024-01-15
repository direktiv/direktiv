import { Antenna } from "lucide-react";
import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const EventsListenerBreadcrumb = () => {
  const namespace = useNamespace();
  const { isEventsListenersPage } = pages.events.useParams();
  const { t } = useTranslation();

  if (!isEventsListenersPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-event-listeners">
        <Link
          to={pages.events.createHref({
            namespace,
            subpage: "eventlisteners",
          })}
        >
          <Antenna aria-hidden="true" />
          {t("components.breadcrumb.eventListeners")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default EventsListenerBreadcrumb;
