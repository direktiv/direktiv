import {
  ActivitySquare,
  Boxes,
  FolderTree,
  GitCompare,
  Layers,
  LucideIcon,
  Radio,
  Settings,
} from "lucide-react";
import { useMatches, useParams, useSearchParams } from "react-router-dom";

import Activities from "~/pages/namespace/Mirror/Activities";
import EventsPage from "~/pages/namespace/Events";
import History from "~/pages/namespace/Events/History";
import InstancesPage from "~/pages/namespace/Instances";
import InstancesPageDetail from "~/pages/namespace/Instances/Detail";
import InstancesPageList from "~/pages/namespace/Instances/List";
import Listeners from "~/pages/namespace/Events/Listeners";
import Logs from "~/pages/namespace/Mirror/Activities/Detail";
import MirrorPage from "~/pages/namespace/Mirror";
import MonitoringPage from "~/pages/namespace/Monitoring";
import type { RouteObject } from "react-router-dom";
import SettingsPage from "~/pages/namespace/Settings";
import TreePage from "~/pages/namespace/Explorer/Tree";
import WorkflowPage from "~/pages/namespace/Explorer/Workflow";
import WorkflowPageActive from "~/pages/namespace/Explorer/Workflow/Active";
import WorkflowPageOverview from "~/pages/namespace/Explorer/Workflow/Overview";
import WorkflowPageRevisions from "~/pages/namespace/Explorer/Workflow/Revisions";
import WorkflowPageSettings from "~/pages/namespace/Explorer/Workflow/Settings";
import { checkHandlerInMatcher as checkHandler } from "./utils";

interface PageBase {
  name: string;
  icon: LucideIcon;
  route: RouteObject;
}

type KeysWithNoPathParams =
  | "monitoring"
  // | "gateway"
  // | "permissions"
  | "services"
  | "settings";

type DefaultPageSetup = Record<
  KeysWithNoPathParams,
  PageBase & { createHref: (params: { namespace: string }) => string }
>;

type ExplorerSubpages =
  | "workflow"
  | "workflow-revisions"
  | "workflow-overview"
  | "workflow-settings";

type ExplorerSubpagesParams =
  | {
      subpage?: Exclude<ExplorerSubpages, "workflow-revisions">;
    }
  // only workflow-revisions has a optional revision param
  | {
      subpage: "workflow-revisions";
      revision?: string;
    };

type ExplorerPageSetup = Record<
  "explorer",
  PageBase & {
    createHref: (
      params: {
        namespace: string;
        path?: string;
        // if no subpage is provided, it opens the tree view
      } & ExplorerSubpagesParams
    ) => string;
    useParams: () => {
      namespace: string | undefined;
      path: string | undefined;
      revision: string | undefined;
      isExplorerPage: boolean;
      isTreePage: boolean;
      isWorkflowPage: boolean;
      isWorkflowActivePage: boolean;
      isWorkflowRevPage: boolean;
      isWorkflowOverviewPage: boolean;
      isWorkflowSettingsPage: boolean;
    };
  }
>;

type InstancesPageSetup = Record<
  "instances",
  PageBase & {
    createHref: (params: { namespace: string; instance?: string }) => string;
    useParams: () => {
      namespace: string | undefined;
      instance: string | undefined;
      isInstancePage: boolean;
      isInstanceListPage: boolean;
      isInstanceDetailPage: boolean;
    };
  }
>;

type EventsPageSetup = Record<
  "events",
  PageBase & {
    createHref: (params: {
      namespace: string;
      subpage?: "eventlisteners";
    }) => string;
    useParams: () => {
      isEventsHistoryPage: boolean;
      isEventsListenersPage: boolean;
    };
  }
>;

type MirrorPageSetup = Record<
  "mirror",
  PageBase & {
    createHref: (params: { namespace: string; activity?: string }) => string;
    useParams: () => {
      activity?: string;
      isMirrorActivityPage: boolean;
      isMirrorLogsPage: boolean;
    };
  }
>;

type MonitoringPageSetup = Record<
  "monitoring",
  PageBase & {
    createHref: (params: { namespace: string }) => string;
    useParams: () => {
      isMonitoringPage: boolean;
    };
  }
>;

type PageType = DefaultPageSetup &
  ExplorerPageSetup &
  InstancesPageSetup &
  EventsPageSetup &
  MonitoringPageSetup &
  MirrorPageSetup;

// these are the direct child pages that live in the /:namespace folder
// the main goal of this abstraction is to make the router as typesafe as
// possible and to globally manage and change the url structure
// entries with no name and icon will not be rendered in the navigation
export const pages: PageType = {
  explorer: {
    name: "components.mainMenu.explorer",
    icon: FolderTree,
    createHref: (params) => {
      let path = "";
      if (params.path) {
        path = params.path.startsWith("/") ? params.path : `/${params.path}`;
      }

      const subfolder: Record<ExplorerSubpages, string> = {
        workflow: "workflow/active",
        "workflow-revisions": "workflow/revisions",
        "workflow-overview": "workflow/overview",
        "workflow-settings": "workflow/settings",
      };

      const searchParams = new URLSearchParams({
        ...(params.subpage === "workflow-revisions" && params.revision
          ? { revision: params.revision }
          : {}),
      });
      const subpage = params.subpage ? subfolder[params.subpage] : "tree";
      return `/${
        params.namespace
      }/explorer/${subpage}${path}?${searchParams.toString()}`;
    },
    useParams: () => {
      const { "*": path, namespace } = useParams();
      const [, , thirdLvl, fourthLvl] = useMatches(); // first level is namespace level
      const [searchParams] = useSearchParams();

      // explorer.useParams() can also be called on pages that are not
      // the explorer page and some params might accidentally match as
      // well (like wildcards). To prevent that we use custom handles that
      // we injected in the route objects
      const isTreePage = checkHandler(thirdLvl, "isTreePage");
      const isWorkflowPage = checkHandler(thirdLvl, "isWorkflowPage");
      const isExplorerPage = isTreePage || isWorkflowPage;
      const isWorkflowActivePage = checkHandler(fourthLvl, "isActivePage");
      const isWorkflowRevPage = checkHandler(fourthLvl, "isRevisionsPage");
      const isWorkflowOverviewPage = checkHandler(fourthLvl, "isOverviewPage");
      const isWorkflowSettingsPage = checkHandler(fourthLvl, "isSettingsPage");

      return {
        path: isExplorerPage ? path : undefined,
        namespace: isExplorerPage ? namespace : undefined,
        isExplorerPage: isTreePage || isWorkflowPage,
        revision: searchParams.get("revision") ?? undefined,
        isTreePage,
        isWorkflowPage,
        isWorkflowActivePage,
        isWorkflowRevPage,
        isWorkflowOverviewPage,
        isWorkflowSettingsPage,
      };
    },
    route: {
      path: "explorer/",
      children: [
        {
          path: "tree/*",
          element: <TreePage />,
          handle: { isTreePage: true },
        },
        {
          path: "workflow/",
          element: <WorkflowPage />,
          handle: { isWorkflowPage: true },
          children: [
            {
              path: "active/*",
              element: <WorkflowPageActive />,
              handle: { isActivePage: true },
            },
            {
              path: "revisions/*",
              element: <WorkflowPageRevisions />,
              handle: { isRevisionsPage: true },
            },
            {
              path: "overview/*",
              element: <WorkflowPageOverview />,
              handle: { isOverviewPage: true },
            },
            {
              path: "settings/*",
              element: <WorkflowPageSettings />,
              handle: { isSettingsPage: true },
            },
          ],
        },
      ],
    },
  },
  monitoring: {
    name: "components.mainMenu.monitoring",
    icon: ActivitySquare,
    createHref: (params) => `/${params.namespace}/monitoring`,
    useParams: () => {
      const [, secondLevel] = useMatches(); // first level is namespace level
      const isMonitoringPage = checkHandler(secondLevel, "isMonitoringPage");
      return { isMonitoringPage };
    },
    route: {
      path: "monitoring",
      element: <MonitoringPage />,
      handle: { isMonitoringPage: true },
    },
  },
  instances: {
    name: "components.mainMenu.instances",
    icon: Boxes,
    createHref: (params) =>
      `/${params.namespace}/instances${
        params.instance ? `/${params.instance}` : ""
      }`,
    useParams: () => {
      const { namespace, instance } = useParams();

      const [, , thirdLvl] = useMatches(); // first level is namespace level

      const isInstanceListPage = checkHandler(thirdLvl, "isInstanceListPage");
      const isInstanceDetailPage = checkHandler(
        thirdLvl,
        "isInstanceDetailPage"
      );

      const isInstancePage = isInstanceListPage || isInstanceDetailPage;

      return {
        namespace: isInstancePage ? namespace : undefined,
        instance: isInstancePage ? instance : undefined,
        isInstancePage,
        isInstanceListPage,
        isInstanceDetailPage,
      };
    },
    route: {
      path: "instances",
      element: <InstancesPage />,
      children: [
        {
          path: "",
          element: <InstancesPageList />,
          handle: { isInstanceListPage: true },
        },
        {
          path: ":instance",
          element: <InstancesPageDetail />,
          handle: { isInstanceDetailPage: true },
        },
      ],
    },
  },
  events: {
    name: "components.mainMenu.events",
    icon: Radio,
    createHref: (params) =>
      `/${params.namespace}/events/${
        params?.subpage === "eventlisteners" ? `listeners` : "history"
      }`,
    useParams: () => {
      const [, , thirdLevel] = useMatches(); // first level is namespace level
      const isEventsHistoryPage = checkHandler(
        thirdLevel,
        "isEventHistoryPage"
      );
      const isEventsListenersPage = checkHandler(
        thirdLevel,
        "isEventListenersPage"
      );
      return { isEventsHistoryPage, isEventsListenersPage };
    },
    route: {
      path: "events",
      element: <EventsPage />,
      children: [
        {
          path: "history",
          element: <History />,
          handle: { isEventHistoryPage: true },
        },
        {
          path: "listeners",
          element: <Listeners />,
          handle: { isEventListenersPage: true },
        },
      ],
    },
  },
  // gateway: {
  //   name: "components.mainMenu.gateway",
  //   icon: Network,
  //   createHref: (params) => `/${params.namespace}/gateway`,
  //   route: {
  //     path: "gateway",
  //     element: <div className="flex flex-col space-y-5 p-10">Gateway</div>,
  //   },
  // },
  // permissions: {
  //   name: "components.mainMenu.permissions",
  //   icon: Users,
  //   createHref: (params) => `/${params.namespace}/permissions`,
  //   route: {
  //     path: "permissions",
  //     element: <div className="flex flex-col space-y-5 p-10">Permissions</div>,
  //   },
  // },
  services: {
    name: "components.mainMenu.services",
    icon: Layers,
    createHref: (params) => `/${params.namespace}/services`,
    route: {
      path: "services",
      element: <div className="flex flex-col space-y-5 p-10">Services</div>,
    },
  },
  mirror: {
    name: "components.mainMenu.mirror",
    icon: GitCompare,
    createHref: (params) =>
      `/${params.namespace}/mirror/activities/${
        params?.activity ? params.activity : ""
      }`,
    useParams: () => {
      const { activity } = useParams();
      const [, , thirdLevel] = useMatches(); // first level is namespace level
      const isMirrorActivityPage = checkHandler(
        thirdLevel,
        "isMirrorActivityPage"
      );
      const isMirrorLogsPage = checkHandler(thirdLevel, "isMirrorLogsPage");
      return {
        isMirrorActivityPage,
        isMirrorLogsPage,
        activity: isMirrorLogsPage ? activity : undefined,
      };
    },
    route: {
      path: "mirror",
      element: <MirrorPage />,
      children: [
        {
          path: "activities",
          element: <Activities />,
          handle: { isMirrorActivitiesPage: true },
        },
        {
          path: "activities/:activity",
          element: <Logs />,
          handle: { isMirrorLogsPage: true },
        },
      ],
    },
  },
  settings: {
    name: "components.mainMenu.settings",
    icon: Settings,
    createHref: (params) => `/${params.namespace}/settings`,
    route: {
      path: "settings",
      element: <SettingsPage />,
    },
  },
};
