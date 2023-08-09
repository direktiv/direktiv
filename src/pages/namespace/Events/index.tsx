import { ArrowDownToDot, Radio } from "lucide-react";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";

import History from "./History";
import Listeners from "./Listeners";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const tabs = ["history", "listeners"] as const;

const EventsPage = () => {
  const [tab, setTab] = useState<(typeof tabs)[number]>("history");
  const { t } = useTranslation();

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
      {tab === "history" && <History />}
      {tab === "listeners" && <Listeners />}
    </>
  );
};

export default EventsPage;
