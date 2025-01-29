import { FolderTree, Layers, LucideIcon } from "lucide-react";
import { Link, useRouterState } from "@tanstack/react-router";

import { FC } from "react";
import { createClassNames } from "~/design/NavigationLink";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

type NavigationItem = {
  path: string;
  label: string;
  icon: LucideIcon;
};

const Navigation: FC = () => {
  const namespace = useNamespace();
  const { location } = useRouterState();
  const { t } = useTranslation();

  if (!namespace) return null;

  const navigationItems: NavigationItem[] = [
    {
      path: "/n/$namespace/explorer",
      label: t("components.mainMenu.explorer"),
      icon: FolderTree,
    },
    {
      path: "/n/$namespace/services",
      label: t("components.mainMenu.services"),
      icon: Layers,
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
