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
import { Link, useRouterState } from "@tanstack/react-router";

import { FC } from "react";
import { FileRoutesByTo } from "~/routeTree.gen";
import { createClassNames } from "~/design/NavigationLink";
import { isEnterprise } from "~/config/env/utils";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

type NavigationItem = {
  path: keyof FileRoutesByTo;
  label: string;
  icon: LucideIcon;
};

const Navigation: FC = () => {
  const namespace = useNamespace();
  const { location } = useRouterState();
  const { t } = useTranslation();

  if (!namespace) return null;

  const enableEnterpriseItems = isEnterprise();

  const enterpriseItems: NavigationItem[] = enableEnterpriseItems
    ? [
        {
          path: "/n/$namespace/permissions",
          label: t("components.mainMenu.permissions"),
          icon: BadgeCheck,
        },
      ]
    : [];

  const navigationItems: NavigationItem[] = [
    {
      path: "/n/$namespace/explorer",
      label: t("components.mainMenu.explorer"),
      icon: FolderTree,
    },
    {
      path: "/n/$namespace/monitoring",
      label: t("components.mainMenu.monitoring"),
      icon: ActivitySquare,
    },
    {
      path: "/n/$namespace/instances",
      label: t("components.mainMenu.instances"),
      icon: Boxes,
    },
    {
      path: "/n/$namespace/events/history",
      label: t("components.mainMenu.events"),
      icon: Radio,
    },
    {
      path: "/n/$namespace/gateway/routes",
      label: t("components.mainMenu.gateway"),
      icon: Network,
    },
    {
      path: "/n/$namespace/services",
      label: t("components.mainMenu.services"),
      icon: Layers,
    },
    {
      path: "/n/$namespace/mirror",
      label: t("components.mainMenu.mirror"),
      icon: GitCompare,
    },
    ...enterpriseItems,
    {
      path: "/n/$namespace/settings",
      label: t("components.mainMenu.settings"),
      icon: Settings,
    },
    {
      path: "/n/$namespace/jq",
      label: t("components.mainMenu.jqPlayground"),
      icon: PlaySquare,
    },
  ];

  return (
    <>
      {navigationItems.map((item) => (
        <Link
          key={item.path}
          to={item.path}
          params={{ namespace }}
          className={createClassNames(location.pathname === item.path)}
        >
          <item.icon aria-hidden="true" /> {item.label}
        </Link>
      ))}
    </>
  );
};

export default Navigation;
