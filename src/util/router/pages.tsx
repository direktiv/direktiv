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
import { useMatches, useParams } from "react-router-dom";

import type { RouteObject } from "react-router-dom";
import SettiongsPage from "../../pages/namespace/Settings";
import TreePage from "../../pages/namespace/Explorer/Tree";
import WorkflowPage from "../../pages/namespace/Explorer/Workflow";
import WorkflowPageActive from "../../pages/namespace/Explorer/Workflow/Active";
import WorkflowPageOverview from "../../pages/namespace/Explorer/Workflow/Overview";
import WorkflowPageRevisions from "../../pages/namespace/Explorer/Workflow/Revisions";
import WorkflowPageSettings from "../../pages/namespace/Explorer/Workflow/Settings";
import { z } from "zod";

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

type ExplorerPageSetup = Record<
  "explorer",
  PageBase & {
    createHref: (params: {
      namespace: string;
      subpage?: ExplorerSubpages; // default is the tree view
      path?: string;
    }) => string;
    useParams: () => {
      namespace: string | undefined;
      path: string | undefined;
      isExplorerPage: boolean;
      isTreePage: boolean;
      isWorkflowPage: boolean;
      isWorkflowActivePage: boolean;
      isWorkflowRevisionsPage: boolean;
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

      const subpage = params.subpage ? subfolder[params.subpage] : "tree";

      return `/${params.namespace}/explorer/${subpage}${path}`;
    },
    useParams: () => {
      const { "*": path, namespace } = useParams();
      const [, , thirdLevel, fourthLevel] = useMatches(); // first level is namespace level

      // explorer.useParams() can also be called on pages that are not
      // the explorer page and some params might accidentally match as
      // well (like wildcards). To prevent that we use custom handles that
      // we injected in the route objects
      const isTreePage = z
        .object({
          handle: z.object({
            isTreePage: z.literal(true),
          }),
        })
        .safeParse(thirdLevel).success;

      const isWorkflowPage = z
        .object({
          handle: z.object({
            isWorkflowPage: z.literal(true),
          }),
        })
        .safeParse(thirdLevel).success;

      const isExplorerPage = isTreePage || isWorkflowPage;

      const isWorkflowActivePage = z
        .object({
          handle: z.object({
            isActivePage: z.literal(true),
          }),
        })
        .safeParse(fourthLevel).success;

      const isWorkflowRevisionsPage = z
        .object({
          handle: z.object({
            isRevisionsPage: z.literal(true),
          }),
        })
        .safeParse(fourthLevel).success;

      const isWorkflowOverviewPage = z
        .object({
          handle: z.object({
            isOverviewPage: z.literal(true),
          }),
        })
        .safeParse(fourthLevel).success;

      const isWorkflowSettingsPage = z
        .object({
          handle: z.object({
            isSettingsPage: z.literal(true),
          }),
        })
        .safeParse(fourthLevel).success;

      return {
        path: isExplorerPage ? path : undefined,
        namespace: isExplorerPage ? namespace : undefined,
        isExplorerPage: isTreePage || isWorkflowPage,
        isTreePage,
        isWorkflowPage,
        isWorkflowActivePage,
        isWorkflowRevisionsPage,
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
      element: <SettiongsPage />,
    },
  },
};
