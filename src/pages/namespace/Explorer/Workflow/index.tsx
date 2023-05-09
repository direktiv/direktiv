import { GitCommit, GitMerge, PieChart, Play, Settings } from "lucide-react";
import { Link, Outlet } from "react-router-dom";
import { Tabs, TabsList, TabsTrigger } from "../../../../design/Tabs";

import Button from "../../../../design/Button";
import { FC } from "react";
import { RxChevronDown } from "react-icons/rx";
import { analyzePath } from "../../../../util/router/utils";
import { pages } from "../../../../util/router/pages";
import { useNamespace } from "../../../../util/store/namespace";
import { useTranslation } from "react-i18next";

const Header: FC = () => {
  const { t } = useTranslation();
  const {
    path,
    isWorkflowActivePage,
    isWorkflowRevPage,
    isWorkflowOverviewPage,
    isWorkflowSettingsPage,
  } = pages.explorer.useParams();
  const namespace = useNamespace();
  const { segments } = analyzePath(path);
  const filename = segments[segments.length - 1];

  if (!namespace) return null;

  const tabs = [
    {
      value: "activeRevision",
      active: isWorkflowActivePage,
      icon: <GitCommit aria-hidden="true" />,
      title: t("pages.explorer.workflow.menu.activeRevision"),
      link: pages.explorer.createHref({
        namespace,
        path,
        subpage: "workflow",
      }),
    },
    {
      value: "revisions",
      active: isWorkflowRevPage,
      icon: <GitMerge aria-hidden="true" />,
      title: t("pages.explorer.workflow.menu.revisions"),
      link: pages.explorer.createHref({
        namespace,
        path,
        subpage: "workflow-revisions",
      }),
    },
    {
      value: "overview",
      active: isWorkflowOverviewPage,
      icon: <PieChart aria-hidden="true" />,
      title: t("pages.explorer.workflow.menu.overview"),
      link: pages.explorer.createHref({
        namespace,
        path,
        subpage: "workflow-overview",
      }),
    },
    {
      value: "settings",
      active: isWorkflowSettingsPage,
      icon: <Settings aria-hidden="true" />,
      title: t("pages.explorer.workflow.menu.settings"),
      link: pages.explorer.createHref({
        namespace,
        path,
        subpage: "workflow-settings",
      }),
    },
  ] as const;

  return (
    <>
      <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 pb-0 dark:border-gray-dark-5 dark:bg-gray-dark-1">
        <div className="flex flex-col max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between">
          <h3
            className="flex items-center gap-x-2 font-bold text-primary-500"
            data-testid="workflow-header"
          >
            <Play className="h-5" />
            {filename?.relative}
          </h3>
          <Button variant="primary">
            {t("pages.explorer.workflow.actionsBtn")} <RxChevronDown />
          </Button>
        </div>
        <div>
          <nav className="-mb-px flex space-x-8">
            <Tabs defaultValue={tabs.find((x) => x.active)?.value}>
              <TabsList>
                {tabs.map((tab) => (
                  <TabsTrigger asChild value={tab.value} key={tab.value}>
                    <Link to={tab.link}>
                      {tab.icon}
                      {tab.title}
                    </Link>
                  </TabsTrigger>
                ))}
              </TabsList>
            </Tabs>
          </nav>
        </div>
      </div>

      <Outlet />
    </>
  );
};

export default Header;
