import {
  Code2,
  Layers,
  PieChart,
  Play,
  Settings,
  TerminalSquare,
} from "lucide-react";
import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { Link, Outlet } from "react-router-dom";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import ApiCommands from "./ApiCommands";
import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Card } from "~/design/Card";
import { FC } from "react";
import { NoPermissions } from "~/design/Table";
import RunWorkflow from "./components/RunWorkflow";
import { analyzePath } from "~/util/router/utils";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";

const Header: FC = () => {
  const { t } = useTranslation();
  const {
    path,
    isWorkflowActivePage,
    isWorkflowOverviewPage,
    isWorkflowSettingsPage,
    isWorkflowServicesPage,
  } = pages.explorer.useParams();
  const namespace = useNamespace();
  const { segments } = analyzePath(path);
  const filename = segments[segments.length - 1];

  const {
    isAllowed,
    noPermissionMessage,
    isFetched: isPermissionCheckFetched,
  } = useNodeContent({ path });

  if (!namespace) return null;
  if (!path) return null;

  const tabs = [
    {
      value: "activeRevision",
      active: isWorkflowActivePage,
      icon: <Code2 aria-hidden="true" />,
      title: t("pages.explorer.workflow.menu.fileContent"),
      link: pages.explorer.createHref({
        namespace,
        path,
        subpage: "workflow",
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
      value: "services",
      active: isWorkflowServicesPage,
      icon: <Layers aria-hidden="true" />,
      title: t("pages.explorer.workflow.menu.services"),
      link: pages.explorer.createHref({
        namespace,
        path,
        subpage: "workflow-services",
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

  if (!isPermissionCheckFetched) return null;

  if (isAllowed === false)
    return (
      <Card className="m-5 flex grow">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

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
          <div className="flex gap-x-3">
            <TooltipProvider>
              <ButtonBar>
                <Dialog>
                  <DialogTrigger asChild>
                    <Button icon variant="outline">
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <TerminalSquare />
                        </TooltipTrigger>
                        <TooltipContent>
                          {t("pages.explorer.workflow.apiCommands.tooltip")}
                        </TooltipContent>
                      </Tooltip>
                    </Button>
                  </DialogTrigger>
                  <DialogContent className="sm:max-w-2xl">
                    <ApiCommands namespace={namespace} path={path} />
                  </DialogContent>
                </Dialog>
              </ButtonBar>
            </TooltipProvider>
            <Dialog>
              <DialogTrigger asChild>
                <Button
                  variant="primary"
                  data-testid="workflow-header-btn-run"
                  className="grow"
                >
                  <Play />
                  {t("pages.explorer.workflow.runBtn")}
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-2xl">
                <RunWorkflow path={path} />
              </DialogContent>
            </Dialog>
          </div>
        </div>
        <div>
          <nav className="-mb-px flex space-x-8">
            <Tabs value={tabs.find((x) => x.active)?.value}>
              <TabsList>
                {tabs.map((tab) => (
                  <TabsTrigger
                    asChild
                    value={tab.value}
                    key={tab.value}
                    data-testid={`workflow-tabs-trg-${tab.value}`}
                  >
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
