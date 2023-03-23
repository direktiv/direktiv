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

import ExplorerPage from "../../pages/namespace/Explorer";
import type { RouteObject } from "react-router-dom";
import SettiongsPage from "../../pages/namespace/Settings";
import WorkflowPage from "../../pages/namespace/Workflow";
import { useParams } from "react-router-dom";

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
      const { "*": path, namespace } = useParams(); // problem, wildcard matches on every route
      return { path, namespace };
    },
    route: {
      path: "explorer/*",
      element: <ExplorerPage />,
    },
  },
  workflow: {
    createHref: (params) =>
      `/${params.namespace}/workflow/${params.path ?? ""}`,
    useParams: () => {
      const { "*": path, namespace } = useParams(); // problem, wildcard matches on every route
      return { path, namespace };
    },
    route: {
      path: "workflow/*",
      element: <WorkflowPage />,
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
