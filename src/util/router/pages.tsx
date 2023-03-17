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

import ExplorerPageSetup from "../../pages/namespace/Explorer";
import type { RouteObject } from "react-router-dom";
import SettiongsPage from "../../pages/namespace/Settings";
import { useParams } from "react-router-dom";

interface PageBase {
  name: string;
  icon: React.FC<React.SVGProps<SVGSVGElement>>;
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
  PageBase & { createHref: () => string }
>;
type ExplorerPageSetup = Record<
  "explorer",
  PageBase & {
    createHref: (params?: { directory: string }) => string;
    useParams: () => { directory: string | undefined };
  }
>;

export const pages: DefaultPageSetup & ExplorerPageSetup = {
  explorer: {
    name: "Explorer",
    icon: FolderTree,
    createHref: (params) =>
      `/explorer${params?.directory ? `/${params.directory}` : ""}`,
    useParams: () => {
      const { "*": directory } = useParams();
      return { directory: directory };
    },
    route: {
      path: "explorer/*",
      element: <ExplorerPageSetup />,
    },
  },
  monitoring: {
    name: "Monitoring",
    icon: Bug,
    createHref: () => "/monitoring",
    route: {
      path: "monitoring",
      element: <div>Monitoring</div>,
    },
  },

  instances: {
    name: "Instances",
    icon: Box,
    createHref: () => "/instances",
    route: {
      path: "instances",
      element: <div>Instances</div>,
    },
  },
  events: {
    name: "Events",
    icon: Calendar,
    createHref: () => "/events",
    route: {
      path: "events",
      element: <div>Events</div>,
    },
  },
  gateway: {
    name: "Gateway",
    icon: Network,
    createHref: () => "/gateway",
    route: {
      path: "gateway",
      element: <div>Gateway</div>,
    },
  },
  permissions: {
    name: "Permissions",
    icon: Users,
    createHref: () => "/permissions",
    route: {
      path: "permissions",
      element: <div>Permissions</div>,
    },
  },
  services: {
    name: "Services",
    icon: Layers,
    createHref: () => "services",
    route: {
      path: "services",
      element: <div>Services</div>,
    },
  },
  settings: {
    name: "Settings",
    icon: Settings,
    createHref: () => "/settings",
    route: {
      path: "settings",
      element: <SettiongsPage />,
    },
  },
};
