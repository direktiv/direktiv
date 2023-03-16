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

import ExplorerPage from "../../pages/Explorer";
import type { RouteObject } from "react-router-dom";
import SettiongsPage from "../../pages/Settings";

type Page = {
  name: string;
  // any is okay here, because every page must implement this function depdening on its params
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  createHref: (...params: any) => string;
  icon: React.FC<React.SVGProps<SVGSVGElement>>;
  route: RouteObject;
};

export const pages: Record<string, Page> = {
  explorer: {
    name: "Explorer",
    icon: FolderTree,
    createHref: () => "explorer",
    route: {
      path: "explorer",
      element: <ExplorerPage />,
    },
  },
  monitoring: {
    name: "Monitoring",
    icon: Bug,
    createHref: () => "monitoring",
    route: {
      path: "monitoring",
      element: <div>Monitoring</div>,
    },
  },

  instances: {
    name: "Instances",
    icon: Box,
    createHref: () => "instances",
    route: {
      path: "instances",
      element: <div>Instances</div>,
    },
  },
  events: {
    name: "Events",
    icon: Calendar,
    createHref: () => "events",
    route: {
      path: "events",
      element: <div>Events</div>,
    },
  },
  gateway: {
    name: "Gateway",
    icon: Network,
    createHref: () => "gateway",
    route: {
      path: "gateway",
      element: <div>Gateway</div>,
    },
  },
  permissions: {
    name: "Permissions",
    icon: Users,
    createHref: () => "permissions",
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
    createHref: () => "settings",
    route: {
      path: "settings",
      element: <SettiongsPage />,
    },
  },
} as const;
