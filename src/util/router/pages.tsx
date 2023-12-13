import {
  ActivitySquare,
  BadgeCheck,
  Boxes,
  FolderTree,
  GitCompare,
  Layers,
  LucideIcon,
  Network,
  PlaySquare,
  Radio,
  Settings,
} from "lucide-react";
import { useMatches, useParams, useSearchParams } from "react-router-dom";

import Activities from "~/pages/namespace/Mirror/Activities";
import ConsumerEditorPage from "~/pages/namespace/Explorer/Consumer";
import EndpointEditorPage from "~/pages/namespace/Explorer/Route";
import EventsPage from "~/pages/namespace/Events";
import GatewayConsumersPage from "~/pages/namespace/Gateway/Consumers";
import GatewayPage from "~/pages/namespace/Gateway";
import GatewayRoutesPage from "~/pages/namespace/Gateway/Routes";
import GroupsPage from "~/pages/namespace/Permissions/Groups";
import History from "~/pages/namespace/Events/History";
import InstancesPage from "~/pages/namespace/Instances";
import InstancesPageDetail from "~/pages/namespace/Instances/Detail";
import InstancesPageList from "~/pages/namespace/Instances/List";
import JqPlaygroundPage from "~/pages/namespace/JqPlayground";
import Listeners from "~/pages/namespace/Events/Listeners";
import Logs from "~/pages/namespace/Mirror/Activities/Detail";
import MirrorPage from "~/pages/namespace/Mirror";
import MonitoringPage from "~/pages/namespace/Monitoring";
import PermissionsPage from "~/pages/namespace/Permissions";
import PolicyPage from "~/pages/namespace/Permissions/Policy";
import type { RouteObject } from "react-router-dom";
import ServiceDetailPage from "~/pages/namespace/Services/Detail";
import ServiceEditorPage from "~/pages/namespace/Explorer/Service";
import ServicesListPage from "~/pages/namespace/Services/List";
import ServicesPage from "~/pages/namespace/Services";
import SettingsPage from "~/pages/namespace/Settings";
import TokensPage from "~/pages/namespace/Permissions/Tokens";
import TreePage from "~/pages/namespace/Explorer/Tree";
import WorkflowPage from "~/pages/namespace/Explorer/Workflow";
import WorkflowPageActive from "~/pages/namespace/Explorer/Workflow/Active";
import WorkflowPageOverview from "~/pages/namespace/Explorer/Workflow/Overview";
import WorkflowPageRevisions from "~/pages/namespace/Explorer/Workflow/Revisions";
import WorkflowPageServices from "~/pages/namespace/Explorer/Workflow/Services";
import WorkflowPageSettings from "~/pages/namespace/Explorer/Workflow/Settings";
import { checkHandlerInMatcher as checkHandler } from "./utils";
import env from "~/config/env";

type PageBase = {
  name: string;
  icon: LucideIcon;
  route: RouteObject;
};

type KeysWithNoPathParams = "monitoring" | "settings" | "jqPlayground";

type DefaultPageSetup = Record<
  KeysWithNoPathParams,
  PageBase & { createHref: (params: { namespace: string }) => string }
>;

export type ExplorerSubpages =
  | "workflow"
  | "workflow-revisions"
  | "workflow-overview"
  | "workflow-settings"
  | "workflow-services"
  | "service"
  | "endpoint"
  | "consumer";

type ExplorerSubpagesParams =
  | {
      subpage?: Exclude<
        ExplorerSubpages,
        "workflow-revisions" | "workflow-services"
      >;
    }
  // workflow-revisions has an optional revision param
  | {
      subpage: "workflow-revisions";
      revision?: string;
    }
  // workflow-services must has an optional serviceId param
  | {
      subpage: "workflow-services";
      serviceId?: string;
    };

type ExplorerPageSetup = Record<
  "explorer",
  PageBase & {
    createHref: (
      params: {
        namespace: string;
        path?: string; // if no subpage is provided, it opens the tree view
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
      isWorkflowServicesPage: boolean;
      isServicePage: boolean;
      isEndpointPage: boolean;
      isConsumerPage: boolean;
      serviceId: string | undefined;
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

type ServicesPageSetup = Record<
  "services",
  PageBase & {
    createHref: (params: { namespace: string; service?: string }) => string;
    useParams: () => {
      namespace: string | undefined;
      service: string | undefined;
      isServicePage: boolean;
      isServiceListPage: boolean;
      isServiceDetailPage: boolean;
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
      isMirrorPage: boolean;
      isActivityDetailPage: boolean;
    };
  }
>;

type MonitoringPageSetup = Record<
  "monitoring",
  PageBase & {
    useParams: () => {
      isMonitoringPage: boolean;
    };
  }
>;

type SettingsPageSetup = Record<
  "settings",
  PageBase & {
    useParams: () => {
      isSettingsPage: boolean;
    };
  }
>;

type JqPlaygroundPageSetup = Record<
  "jqPlayground",
  PageBase & {
    useParams: () => {
      isJqPlaygroundPage: boolean;
    };
  }
>;

type GatewayPageSetup = Record<
  "gateway",
  PageBase & {
    createHref: (params: {
      namespace: string;
      subpage?: "consumers";
    }) => string;
    useParams: () => {
      isGatewayPage: boolean;
      isGatewayRoutesPage: boolean;
      isGatewayConsumerPage: boolean;
    };
  }
>;

type PageType = DefaultPageSetup &
  ExplorerPageSetup &
  InstancesPageSetup &
  ServicesPageSetup &
  EventsPageSetup &
  MonitoringPageSetup &
  SettingsPageSetup &
  GatewayPageSetup &
  JqPlaygroundPageSetup &
  MirrorPageSetup;

type PermissionsPageSetup = Partial<
  Record<
    "permissions",
    PageBase & {
      createHref: (params: {
        namespace: string;
        subpage?: "tokens" | "groups"; // policy is the default page
      }) => string;
      useParams: () => {
        isPermissionsPage: boolean;
        isPermissionsPolicyPage: boolean;
        isPermissionsTokenPage: boolean;
        isPermissionsGroupPage: boolean;
      };
    }
  >
>;

type EnterprisePageType = PermissionsPageSetup;

export const enterprisePages: EnterprisePageType = env.VITE_IS_ENTERPRISE
  ? {
      permissions: {
        name: "components.mainMenu.permissions",
        icon: BadgeCheck,
        createHref: (params) => {
          let subpage = "";
          if (params.subpage === "groups") {
            subpage = "/groups";
          }
          if (params.subpage === "tokens") {
            subpage = "/tokens";
          }
          return `/${params.namespace}/permissions${subpage}`;
        },
        useParams: () => {
          const [, secondLevel, thirdLevel] = useMatches(); // first level is namespace level
          const isPermissionsPage = checkHandler(
            secondLevel,
            "isPermissionsPage"
          );
          const isPermissionsPolicyPage = checkHandler(
            thirdLevel,
            "isPermissionsPolicyPage"
          );
          const isPermissionsTokenPage = checkHandler(
            thirdLevel,
            "isPermissionsTokenPage"
          );
          const isPermissionsGroupPage = checkHandler(
            thirdLevel,
            "isPermissionsGroupPage"
          );

          return {
            isPermissionsPage,
            isPermissionsPolicyPage,
            isPermissionsTokenPage,
            isPermissionsGroupPage,
          };
        },
        route: {
          path: "permissions",
          element: <PermissionsPage />,
          handle: { permissions: true, isPermissionsPage: true },
          children: [
            {
              path: "",
              element: <PolicyPage />,
              handle: { isPermissionsPolicyPage: true },
            },
            {
              path: "tokens",
              element: <TokensPage />,
              handle: { isPermissionsTokenPage: true },
            },
            {
              path: "groups",
              element: <GroupsPage />,
              handle: { isPermissionsGroupPage: true },
            },
          ],
        },
      },
    }
  : {};

// these are the direct child pages that live in the /:namespace folder
// the main goal of this abstraction is to make the router as typesafe as
// possible and to globally manage and change the url structure
// entries with no name and icon will not be rendered in the navigation
export const pages: PageType & EnterprisePageType = {
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
        "workflow-services": "workflow/services",
        endpoint: "endpoint",
        consumer: "consumer",
        service: "service",
      };

      let searchParamsObj;

      if (params.subpage === "workflow-revisions" && params.revision) {
        searchParamsObj = { revision: params.revision };
      }

      if (params.subpage === "workflow-services" && params.serviceId) {
        searchParamsObj = {
          serviceId: params.serviceId,
        };
      }

      const searchParams = new URLSearchParams(searchParamsObj);

      const subpage = params.subpage ? subfolder[params.subpage] : "tree";

      const searchParamsString = searchParams.toString();
      const urlParams = searchParamsString ? `?${searchParamsString}` : "";

      return `/${params.namespace}/explorer/${subpage}${path}${urlParams}`;
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
      const isServicePage = checkHandler(thirdLvl, "isServicePage");
      const isEndpointPage = checkHandler(thirdLvl, "isEndpointPage");
      const isConsumerPage = checkHandler(thirdLvl, "isConsumerPage");
      const isExplorerPage =
        isTreePage ||
        isWorkflowPage ||
        isServicePage ||
        isEndpointPage ||
        isConsumerPage;
      const isWorkflowActivePage = checkHandler(fourthLvl, "isActivePage");
      const isWorkflowRevPage = checkHandler(fourthLvl, "isRevisionsPage");
      const isWorkflowOverviewPage = checkHandler(fourthLvl, "isOverviewPage");
      const isWorkflowSettingsPage = checkHandler(fourthLvl, "isSettingsPage");
      const isWorkflowServicesPage = checkHandler(fourthLvl, "isServicesPage");

      return {
        path: isExplorerPage ? path : undefined,
        namespace: isExplorerPage ? namespace : undefined,
        isExplorerPage,
        revision: searchParams.get("revision") ?? undefined,
        isTreePage,
        isWorkflowPage,
        isWorkflowActivePage,
        isWorkflowRevPage,
        isWorkflowOverviewPage,
        isWorkflowSettingsPage,
        isWorkflowServicesPage,
        isServicePage,
        isEndpointPage,
        isConsumerPage,
        serviceId: searchParams.get("serviceId") ?? undefined,
      };
    },
    route: {
      path: "explorer/",
      handle: { explorer: true },
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
            {
              path: "services/*",
              element: <WorkflowPageServices />,
              handle: { isServicesPage: true },
            },
          ],
        },
        {
          path: "service/*",
          element: <ServiceEditorPage />,
          handle: { isServicePage: true },
        },
        {
          path: "endpoint/*",
          element: <EndpointEditorPage />,
          handle: { isEndpointPage: true },
        },
        {
          path: "consumer/*",
          element: <ConsumerEditorPage />,
          handle: { isConsumerPage: true },
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
      handle: { monitoring: true, isMonitoringPage: true },
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
      handle: { instances: true },
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
      handle: { events: true },
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
  gateway: {
    name: "components.mainMenu.gateway",
    icon: Network,
    createHref: (params) =>
      `/${params.namespace}/gateway/${
        params?.subpage === "consumers" ? `consumers` : "routes"
      }`,
    useParams: () => {
      const [, secondLevel, thirdLevel] = useMatches(); // first level is namespace level
      const isGatewayPage = checkHandler(secondLevel, "isGatewayPage");
      const isGatewayRoutesPage = checkHandler(
        thirdLevel,
        "isGatewayRoutesPage"
      );
      const isGatewayConsumerPage = checkHandler(
        thirdLevel,
        "isGatewayConsumerPage"
      );
      return {
        isGatewayPage,
        isGatewayRoutesPage,
        isGatewayConsumerPage,
      };
    },
    route: {
      path: "gateway",
      element: <GatewayPage />,
      handle: { gateway: true, isGatewayPage: true },
      children: [
        {
          path: "routes",
          element: <GatewayRoutesPage />,
          handle: { isGatewayRoutesPage: true },
        },
        {
          path: "consumers",
          element: <GatewayConsumersPage />,
          handle: { isGatewayConsumerPage: true },
        },
      ],
    },
  },
  services: {
    name: "components.mainMenu.services",
    icon: Layers,
    createHref: (params) =>
      `/${params.namespace}/services${
        params.service ? `/${params.service}` : ""
      }`,
    useParams: () => {
      const { namespace, service } = useParams();

      const [, , thirdLvl] = useMatches(); // first level is namespace level

      const isServiceListPage = checkHandler(thirdLvl, "isServiceListPage");
      const isServiceDetailPage = checkHandler(thirdLvl, "isServiceDetailPage");
      const isServicePage = isServiceListPage || isServiceDetailPage;

      return {
        namespace: isServicePage ? namespace : undefined,
        service: isServiceDetailPage ? service : undefined,
        isServicePage,
        isServiceListPage,
        isServiceDetailPage,
      };
    },

    route: {
      path: "services",
      element: <ServicesPage />,
      handle: { services: true },
      children: [
        {
          path: "",
          element: <ServicesListPage />,
          handle: { isServiceListPage: true },
        },
        {
          path: ":service",
          element: <ServiceDetailPage />,
          handle: { isServiceDetailPage: true },
        },
      ],
    },
  },
  mirror: {
    name: "components.mainMenu.mirror",
    icon: GitCompare,
    createHref: (params) =>
      `/${params.namespace}/mirror/${
        params?.activity ? `logs/${params.activity}` : ""
      }`,
    useParams: () => {
      const { activity } = useParams();
      const [, secondLevel, thirdLevel] = useMatches(); // first level is namespace level
      const isMirrorPage = checkHandler(secondLevel, "isMirrorPage");
      const isActivityDetailPage = checkHandler(thirdLevel, "isMirrorLogsPage");
      return {
        isMirrorPage,
        isActivityDetailPage,
        activity: isActivityDetailPage ? activity : undefined,
      };
    },
    route: {
      path: "mirror",
      element: <MirrorPage />,
      handle: { mirror: true, isMirrorPage: true },
      children: [
        {
          path: "",
          element: <Activities />,
          handle: { isMirrorActivitiesPage: true },
        },
        {
          path: "logs/:activity",
          element: <Logs />,
          handle: { isMirrorLogsPage: true },
        },
      ],
    },
  },
  ...enterprisePages,
  settings: {
    name: "components.mainMenu.settings",
    icon: Settings,
    createHref: (params) => `/${params.namespace}/settings`,
    useParams: () => {
      const [, secondLevel] = useMatches(); // first level is namespace level
      const isSettingsPage = checkHandler(secondLevel, "isSettingsPage");
      return { isSettingsPage };
    },
    route: {
      path: "settings",
      element: <SettingsPage />,
      handle: { settings: true, isSettingsPage: true },
    },
  },
  jqPlayground: {
    name: "components.mainMenu.jqPlayground",
    icon: PlaySquare,
    createHref: (params) => `/${params.namespace}/jq`,
    useParams: () => {
      const [, secondLevel] = useMatches(); // first level is namespace level
      const isJqPlaygroundPage = checkHandler(
        secondLevel,
        "isJqPlaygroundPage"
      );
      return { isJqPlaygroundPage };
    },
    route: {
      path: "jq",
      element: <JqPlaygroundPage />,
      handle: { jqPlayground: true, isJqPlaygroundPage: true },
    },
  },
};
