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

import ExplorerPage from "../../pages/namespace/Explorer";
import type { RouteObject } from "react-router-dom";
import SettiongsPage from "../../pages/namespace/Settings";
import WorkflowPage from "../../pages/namespace/Workflow";
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

type ExplorerPageSetup = Record<
  "explorer",
  PageBase & {
    createHref: (params: { namespace: string; path?: string }) => string;
    useParams: () => {
      namespace: string | undefined;
      path: string | undefined;
    };
  }
>;

type WorkflowPageSetup = Record<
  "workflow",
  PageBase & {
    createHref: (params: { namespace: string; path?: string }) => string;
    useParams: () => {
      namespace: string | undefined;
      path: string | undefined;
    };
  }
>;

// these are the direct child pages that live in the /:namespace folder
// the main goal of this abstraction is to make the router as typesafe as
// possible and to globally manage and change the url structure
// entries with no name and icon will not be rendered in the navigation
export const pages: DefaultPageSetup & ExplorerPageSetup & WorkflowPageSetup = {
  explorer: {
    name: "Explorer",
    icon: FolderTree,
    createHref: (params) =>
      `/${params.namespace}/explorer${`/${params.path ?? ""}`}`,
    useParams: () => {
      const { "*": path, namespace } = useParams();
      const [, secondLevel] = useMatches(); // first level is namespace level
      // explorer.useParams() can also be called on pages
      // that are not the explorer page and some params
      // might accidentally match as well. To prevent that
      // we use the custom handle we injected in the route
      const isExplorer = z
        .object({
          handle: z.object({
            isExplorerPage: z.literal(true),
          }),
        })
        .safeParse(secondLevel).success;
      return {
        path: isExplorer ? path : undefined,
        namespace: isExplorer ? namespace : undefined,
      };
    },
    route: {
      path: "explorer/*",
      element: <ExplorerPage />,
      handle: { isExplorerPage: true },
    },
  },
  workflow: {
    createHref: (params) =>
      `/${params.namespace}/workflow/${params.path ?? ""}`,
    useParams: () => {
      const { "*": path, namespace } = useParams();
      const [, secondLevel] = useMatches();
      const isWorkflowPage = z
        .object({
          handle: z.object({
            isWorkflowPage: z.literal(true),
          }),
        })
        .safeParse(secondLevel).success;

      return {
        path: isWorkflowPage ? path : undefined,
        namespace: isWorkflowPage ? namespace : undefined,
      };
    },
    route: {
      path: "workflow/*",
      element: <WorkflowPage />,
      handle: { isWorkflowPage: true },
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
} as const;
