import { ArrowDownToDot, Radio } from "lucide-react";
import { FiltersObj, useEventsStream } from "~/api/events/query/get";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";

import EventListeners from "../Listeners";
import EventsList from "./EventsList";
import { useState } from "react";
import { useTranslation } from "react-i18next";

export const itemsPerPage = 10;

const tabs = ["history", "listeners"] as const;

const EventsListWrapper = () => {
  const [filters, setFilters] = useState<FiltersObj>({});
  const [offset, setOffset] = useState(0);
  const [tab, setTab] = useState<(typeof tabs)[number]>("history");
  const { t } = useTranslation();

  useEventsStream({ limit: itemsPerPage, offset, filters });

  return (
    <>
      <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 pb-0 dark:border-gray-dark-5 dark:bg-gray-dark-1">
        <Tabs defaultValue="history">
          <TabsList>
            <TabsTrigger
              value="history"
              asChild
              onClick={() => setTab("history")}
            >
              <a href="#">
                <Radio aria-hidden="true" />
                {t("pages.events.tabs.history")}
              </a>
            </TabsTrigger>
            <TabsTrigger
              value="listeners"
              asChild
              onClick={() => setTab("listeners")}
            >
              <a href="#">
                <ArrowDownToDot aria-hidden="true" />
                {t("pages.events.tabs.listeners")}
              </a>
            </TabsTrigger>
          </TabsList>
        </Tabs>
      </div>
      {tab === "history" && (
        <EventsList
          filters={filters}
          setFilters={setFilters}
          offset={offset}
          setOffset={setOffset}
        />
      )}
      {tab === "listeners" && <EventListeners />}
    </>
  );
};

export default EventsListWrapper;
