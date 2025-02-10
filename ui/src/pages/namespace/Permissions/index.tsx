import { KeyRound, Users } from "lucide-react";
import { Link, Outlet, useMatch } from "@tanstack/react-router";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";

import { useTranslation } from "react-i18next";

const PermissionsPage = () => {
  const { t } = useTranslation();

  const isPermissionsRolesPage = useMatch({
    from: "/n/$namespace/permissions/",
    shouldThrow: false,
  });
  const isPermissionsTokensPage = useMatch({
    from: "/n/$namespace/permissions/tokens",
    shouldThrow: false,
  });

  const tabs = [
    {
      value: "roles",
      active: isPermissionsRolesPage,
      icon: <Users aria-hidden="true" />,
      title: t("pages.permissions.tabs.roles"),
      link: "/n/$namespace/permissions",
    },
    {
      value: "tokens",
      active: isPermissionsTokensPage,
      icon: <KeyRound aria-hidden="true" />,
      title: t("pages.permissions.tabs.tokens"),
      link: "/n/$namespace/permissions/tokens",
    },
  ] as const;

  return (
    <>
      <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 pb-0 dark:border-gray-dark-5 dark:bg-gray-dark-1">
        <Tabs value={tabs.find((tab) => tab.active)?.value}>
          <TabsList>
            {tabs.map((tab) => (
              <TabsTrigger
                asChild
                value={tab.value}
                key={tab.value}
                data-testid={`event-tabs-trg-${tab.value}`}
              >
                <Link to={tab.link} from="/n/$namespace">
                  {tab.icon}
                  {tab.title}
                </Link>
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      </div>
      <Outlet />
    </>
  );
};

export default PermissionsPage;
