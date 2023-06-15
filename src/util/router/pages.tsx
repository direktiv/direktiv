import {
  Box,
  Bug,
  Calendar,
  FolderTree,
  Layers,
  Network,
  Settings,
  Users,
} from "lucide-react";
import { useMatches, useParams, useSearchParams } from "react-router-dom";

import React from "react";
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
  name?: string;
  icon?: React.FC<React.SVGProps<SVGSVGElement>>;
  route: RouteObject;
}

type KeysWithNoPathParams =
  | "monitoring"
  | "instances"
  | "events"
  | "gateway"
  | "permissions"
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

// these are the direct child pages that live in the /:namespace folder
// the main goal of this abstraction is to make the router as typesafe as
// possible and to globally manage and change the url structure
// entries with no name and icon will not be rendered in the navigation
export const pages: DefaultPageSetup & ExplorerPageSetup = {
  explorer: {
    name: "Explorer",
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
    name: "Monitoring",
    icon: Bug,
    createHref: (params) => `/${params.namespace}/monitoring`,
    route: {
      path: "monitoring",
      element: <div className="flex flex-col space-y-5 p-10">Monitoring</div>,
    },
  },
  instances: {
    name: "Instances",
    icon: Box,
    createHref: (params) => `/${params.namespace}/instances`,
    route: {
      path: "instances",
      element: <div className="flex flex-col space-y-5 p-10">Instances</div>,
    },
  },
  events: {
    name: "Events",
    icon: Calendar,
    createHref: (params) => `/${params.namespace}/events`,
    route: {
      path: "events",
      element: <div className="flex flex-col space-y-5 p-10">Events</div>,
    },
  },
  gateway: {
    name: "Gateway",
    icon: Network,
    createHref: (params) => `/${params.namespace}/gateway`,
    route: {
      path: "gateway",
      element: <div className="flex flex-col space-y-5 p-10">Gateway</div>,
    },
  },
  permissions: {
    name: "Permissions",
    icon: Users,
    createHref: (params) => `/${params.namespace}/permissions`,
    route: {
      path: "permissions",
      element: <div className="flex flex-col space-y-5 p-10">Permissions</div>,
    },
  },
  services: {
    name: "Services",
    icon: Layers,
    createHref: (params) => `/${params.namespace}/services`,
    route: {
      path: "services",
      element: <div className="flex flex-col space-y-5 p-10">Services</div>,
    },
  },
  settings: {
    name: "Settings",
    icon: Settings,
    createHref: (params) => `/${params.namespace}/settings`,
    route: {
      path: "settings",
      element: <SettingsPage />,
    },
  },
};
