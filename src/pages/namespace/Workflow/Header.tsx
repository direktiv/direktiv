import { GitCommit, GitMerge, PieChart, Play, Settings } from "lucide-react";
import { Tabs, TabsList, TabsTrigger } from "../../../design/Tabs";

import Button from "../../../design/Button";
import { FC } from "react";
import { RxChevronDown } from "react-icons/rx";
import { analyzePath } from "../../../util/router/utils";
import { pages } from "../../../util/router/pages";
import { useTranslation } from "react-i18next";

const Header: FC = () => {
  const { t } = useTranslation();
  const { path } = pages.workflow.useParams();
  const { segments } = analyzePath(path);
  const filename = segments[segments.length - 1];

  const tabs = [
    {
      value: "activeRevision",
      icon: <GitCommit aria-hidden="true" />,
      title: t("pages.explorer.workflow.menu.activeRevision"),
    },
    {
      value: "revisions",
      icon: <GitMerge aria-hidden="true" />,
      title: t("pages.explorer.workflow.menu.revisions"),
    },
    {
      value: "overview",
      icon: <PieChart aria-hidden="true" />,
      title: t("pages.explorer.workflow.menu.overview"),
    },
    {
      value: "settings",
      icon: <Settings aria-hidden="true" />,
      title: t("pages.explorer.workflow.menu.settings"),
    },
  ] as const;

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-2 p-5 pb-0 dark:border-gray-dark-5">
      <div className="flex flex-col max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between">
        <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
          <Play className="h-5" />
          {filename?.relative}
        </h3>
        <Button variant="primary">
          {t("pages.explorer.workflow.actionsBtn")} <RxChevronDown />
        </Button>
      </div>
      <div>
        <nav className="-mb-px flex space-x-8">
          <Tabs>
            <TabsList>
              {tabs.map((tab) => (
                <TabsTrigger asChild value={tab.value} key={tab.value}>
                  <a>
                    {tab.icon}
                    {tab.title}
                  </a>
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
        </nav>
      </div>
    </div>
  );
};

export default Header;
