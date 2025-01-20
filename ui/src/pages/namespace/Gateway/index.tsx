import { Info, Users, Workflow } from "lucide-react";
import { Link, Outlet } from "react-router-dom";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";

import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const GatewayPage = () => {
  const pages = usePages();
  const namespace = useNamespace();
  const { t } = useTranslation();
  const {
    isGatewayRoutesPage,
    isGatewayConsumerPage,
    isGatewayRoutesDetailPage,
    isGatewayInfoPage,
  } = pages.gateway.useParams();

  if (!namespace) return null;

  const tabs = [
    {
      value: "info",
      active: isGatewayInfoPage,
      icon: <Info aria-hidden="true" />,
      title: "Info",
      link: pages.gateway.createHref({
        namespace,
        subpage: "GatewayInfoPage",
      }),
    },
    {
      value: "endpoints",
      active: isGatewayRoutesPage,
      icon: <Workflow aria-hidden="true" />,
      title: t("pages.gateway.tabs.routes"),
      link: pages.gateway.createHref({
        namespace,
      }),
    },
    {
      value: "consumers",
      active: isGatewayConsumerPage,
      icon: <Users aria-hidden="true" />,
      title: t("pages.gateway.tabs.consumers"),
      link: pages.gateway.createHref({
        namespace,
        subpage: "consumers",
      }),
    },
  ] as const;

  return (
    <>
      {!isGatewayRoutesDetailPage && (
        <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 pb-0 dark:border-gray-dark-5 dark:bg-gray-dark-1">
          <Tabs value={tabs.find((tab) => tab.active)?.value}>
            <TabsList>
              {tabs.map((tab) => (
                <TabsTrigger asChild value={tab.value} key={tab.value}>
                  <Link to={tab.link}>
                    {tab.icon}
                    {tab.title}
                  </Link>
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
        </div>
      )}
      <Outlet />
    </>
  );
};

export default GatewayPage;
