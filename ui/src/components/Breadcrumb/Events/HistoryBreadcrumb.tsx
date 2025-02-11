import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Radio } from "lucide-react";
import { useTranslation } from "react-i18next";

const EventsListenerBreadcrumb = () => {
  const isEventsHistoryPage = useMatch({
    from: "/n/$namespace/events/history",
    shouldThrow: false,
  });

  const { t } = useTranslation();

  if (!isEventsHistoryPage) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-event-history">
        <Link to="/n/$namespace/events/history" from="/n/$namespace">
          <Radio aria-hidden="true" />
          {t("components.breadcrumb.eventHistory")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default EventsListenerBreadcrumb;
