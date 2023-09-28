import { FileCheck, KeyRound, Users } from "lucide-react";
import { Link, Outlet } from "react-router-dom";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";

import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const PermissionsPage = () => {
  const permissions = pages.permissions;

  const namespace = useNamespace();
  const { t } = useTranslation();

  if (!permissions) return null;
  if (!namespace) return null;

  const {
    isPermissionsGroupPage,
    isPermissionsPolicyPage,
    isPermissionsTokenPage,
  } = permissions.useParams();

  const tabs = [
    {
      value: "policy",
      active: isPermissionsPolicyPage,
      icon: <FileCheck aria-hidden="true" />,
      title: t("pages.permissions.tabs.policy"),
      link: permissions.createHref({
        namespace,
      }),
    },
    {
      value: "groups",
      active: isPermissionsGroupPage,
      icon: <Users aria-hidden="true" />,
      title: t("pages.permissions.tabs.groups"),
      link: permissions.createHref({
        namespace,
        subpage: "groups",
      }),
    },
    {
      value: "tokens",
      active: isPermissionsTokenPage,
      icon: <KeyRound aria-hidden="true" />,
      title: t("pages.permissions.tabs.tokens"),
      link: permissions.createHref({
        namespace,
        subpage: "tokens",
      }),
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
                <Link to={tab.link}>
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
