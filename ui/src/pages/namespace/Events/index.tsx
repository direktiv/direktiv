import { Antenna, Radio } from "lucide-react";
import { Link, Outlet, useMatch } from "@tanstack/react-router";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";

import { useTranslation } from "react-i18next";

const EventsPage = () => {
  const { t } = useTranslation();
  const isEventsHistoryPage = useMatch({
    from: "/n/$namespace/events/history",
    shouldThrow: false,
  });
  const isEventsListenersPage = useMatch({
    from: "/n/$namespace/events/listeners",
    shouldThrow: false,
  });

  const tabs = [
    {
      value: "history",
      active: isEventsHistoryPage,
      icon: <Radio aria-hidden="true" />,
      title: t("pages.events.tabs.history"),
      link: "/n/$namespace/events/history",
    },
    {
      value: "listeners",
      active: isEventsListenersPage,
      icon: <Antenna aria-hidden="true" />,
      title: t("pages.events.tabs.listeners"),
      link: "/n/$namespace/events/listeners",
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
                <Link to={tab.link} from="/n/$namespace">
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
