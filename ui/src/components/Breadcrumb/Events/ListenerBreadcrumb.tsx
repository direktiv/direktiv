import { Link, useMatch } from "@tanstack/react-router";

import { Antenna } from "lucide-react";
import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { useTranslation } from "react-i18next";

const EventsListenerBreadcrumb = () => {
  const isEventsListenersPage = useMatch({
    from: "/n/$namespace/events/listeners",
    shouldThrow: false,
  });
  const { t } = useTranslation();

  if (!isEventsListenersPage) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-event-listeners">
        <Link to="/n/$namespace/events/history" from="/n/$namespace">
          <Antenna aria-hidden="true" />
          {t("components.breadcrumb.eventListeners")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default EventsListenerBreadcrumb;
