import { BookOpen, Users, Workflow, X } from "lucide-react";
import { Dialog, DialogContent } from "~/design/Dialog";
import { Link, Outlet } from "react-router-dom";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";

import Button from "~/design/Button";
import { RapiDoc } from "~/design/RapiDoc";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const GatewayPage = () => {
  const pages = usePages();
  const namespace = useNamespace();
  const { t } = useTranslation();
  const {
    isGatewayRoutesPage,
    isGatewayConsumerPage,
    isGatewayRoutesDetailPage,
  } = pages.gateway.useParams();

  const spec = "http://localhost:8888/openapi.yaml";

  const [showDocs, setShowDocs] = useState(false);

  if (!namespace) return null;

  const tabs = [
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
        <div className="space-y-5 border-b border-gray-5  p-5 pb-0 ">
          <Button
            onClick={() => setShowDocs(!showDocs)}
            className="absolute right-5 whitespace-nowrap"
            icon
            variant="primary"
          >
            <BookOpen className="mr-2 h-4 w-4" />
            API Documentation
          </Button>
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

      <Dialog open={showDocs} onOpenChange={setShowDocs}>
        <DialogContent className="h-[90vh] min-w-[90vw]  pt-10">
          <Button
            onClick={() => setShowDocs(false)}
            className="absolute right-4 top-4"
            variant="ghost"
            size="sm"
          >
            <X className="h-4 w-4" />
          </Button>
          <RapiDoc spec={spec} />
        </DialogContent>
      </Dialog>

      <Outlet />
    </>
  );
};

export default GatewayPage;
