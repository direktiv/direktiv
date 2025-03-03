import { BookOpen, ScrollText, Users, Workflow } from "lucide-react";
import { Link, Outlet, useMatch } from "@tanstack/react-router";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";

import { useTranslation } from "react-i18next";

const GatewayPage = () => {
  const { t } = useTranslation();
  const isGatewayInfoPage = useMatch({
    from: "/n/$namespace/gateway/gatewayInfo",
    shouldThrow: false,
  });
  const isGatewayRoutesPage = useMatch({
    from: "/n/$namespace/gateway/routes/",
    shouldThrow: false,
  });
  const isGatewayConsumerPage = useMatch({
    from: "/n/$namespace/gateway/consumers",
    shouldThrow: false,
  });
  const isGatewayRoutesDetailPage = useMatch({
    from: "/n/$namespace/gateway/routes/$",
    shouldThrow: false,
  });
  const isGatewayOpenapiDocPage = useMatch({
    from: "/n/$namespace/gateway/openapiDoc",
    shouldThrow: false,
  });

  const tabs = [
    {
      value: "info",
      active: isGatewayInfoPage,
      icon: <BookOpen aria-hidden="true" />,
      title: t("pages.gateway.tabs.info"),
      link: "/n/$namespace/gateway/gatewayInfo",
    },
    {
      value: "endpoints",
      active: isGatewayRoutesPage,
      icon: <Workflow aria-hidden="true" />,
      title: t("pages.gateway.tabs.routes"),
      link: "/n/$namespace/gateway/routes",
    },
    {
      value: "consumers",
      active: isGatewayConsumerPage,
      icon: <Users aria-hidden="true" />,
      title: t("pages.gateway.tabs.consumers"),
      link: "/n/$namespace/gateway/consumers",
    },
    {
      value: "openapiDoc",
      active: isGatewayOpenapiDocPage,
      icon: <ScrollText aria-hidden="true" />,
      title: t("pages.gateway.tabs.documentation"),
      link: "/n/$namespace/gateway/openapiDoc",
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
                  <Link to={tab.link} from="/n/$namespace">
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
