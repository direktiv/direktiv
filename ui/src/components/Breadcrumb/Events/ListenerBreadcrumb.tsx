import { Link, useMatch } from "@tanstack/react-router";

import { Antenna } from "lucide-react";
import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const EventsListenerBreadcrumb = () => {
  const namespace = useNamespace();
  const isEventsListenersPage = useMatch({
    from: "/n/$namespace/events/listeners",
    shouldThrow: false,
  });
  const { t } = useTranslation();

  if (!isEventsListenersPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-event-listeners">
        <Link to="/n/$namespace/events/history" params={{ namespace }}>
          <Antenna aria-hidden="true" />
          {t("components.breadcrumb.eventListeners")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default EventsListenerBreadcrumb;
