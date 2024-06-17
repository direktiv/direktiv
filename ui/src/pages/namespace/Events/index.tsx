import { Antenna, Radio } from "lucide-react";
import { Link, Outlet } from "react-router-dom";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";

import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const EventsPage = () => {
  const pages = usePages();
  const namespace = useNamespace();
  const { t } = useTranslation();
  const { isEventsHistoryPage, isEventsListenersPage } =
    pages.events.useParams();

  if (!namespace) return null;

  const tabs = [
    {
      value: "history",
      active: isEventsHistoryPage,
      icon: <Radio aria-hidden="true" />,
      title: t("pages.events.tabs.history"),
      link: pages.events.createHref({
        namespace,
      }),
    },
    {
      value: "listeners",
      active: isEventsListenersPage,
      icon: <Antenna aria-hidden="true" />,
      title: t("pages.events.tabs.listeners"),
      link: pages.events.createHref({
        namespace,
        subpage: "eventlisteners",
      }),
    },
  ] as const;

  return (
    <>
      <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 pb-0 dark:border-gray-dark-5 dark:bg-gray-dark-1">
        <Tabs value={tabs.find((tab) => tab.active)?.value}>
          <TabsList>
            {tabs.map((tab) => (
              <TabsTrigger
                asChild
                value={tab.value}
                key={tab.value}
                data-testid={`event-tabs-trg-${tab.value}`}
              >
                <Link to={tab.link}>
                  {tab.icon}
                  {tab.title}
                </Link>
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      </div>
      <Outlet />
    </>
  );
};

export default EventsPage;
