import {
  PodLogsSubscriber,
  usePodLogs,
} from "~/api/services/query/revision/pods/getLogs";
import {
  PodsSubscriber,
  usePods,
} from "~/api/services/query/revision/pods/getAll";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Badge from "~/design/Badge";
import { ButtonBar } from "~/design/ButtonBar";
import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import { PodsListSchemaType } from "~/api/services/schema";
import ScrollContainer from "./ScrollContainer";
import { ScrollText } from "lucide-react";
import { podStatusToBadgeVariant } from "../../../components/utils";
import { useState } from "react";
import { useTranslation } from "react-i18next";

export const Pods = ({
  service,
  revision,
}: {
  service: string;
  revision: string;
}) => {
  const { data: podsList, isSuccess } = usePods({ revision, service });
  if (!isSuccess) return null;

  return (
    <>
      <PodsSubscriber revision={revision} service={service} />
      <PodsWithData pods={podsList.pods} />
    </>
  );
};

export const PodsWithData = ({
  pods,
}: {
  pods: PodsListSchemaType["pods"];
}) => {
  const { t } = useTranslation();
  const [selectedTab, setSelectedTab] = useState(pods[0]?.name ?? "");
  const { data: logData } = usePodLogs({
    name: selectedTab,
  });

  const pod = pods.find((pod) => pod.name === selectedTab);

  const logs = logData?.data.split("\n") ?? [];

  if (!pod) return null;

  return (
    <div>
      <PodLogsSubscriber name={selectedTab} />
      <Card className="gap-3 border p-5">
        <div>
          <div className="mb-5 flex flex-col gap-5 sm:flex-row">
            <h3 className="flex grow items-center gap-x-2 font-medium">
              <ScrollText className="h-5" />
              {t("pages.services.revision.detail.logs.title", {
                name: pod.name,
              })}
              <Badge
                variant={podStatusToBadgeVariant(pod.status)}
                className="font-normal"
              >
                {pod.status}
              </Badge>
            </h3>

            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <CopyButton
                    value={logData?.data ?? ""}
                    buttonProps={{
                      variant: "outline",
                      size: "sm",
                      disabled: !logData?.data,
                    }}
                  />
                </TooltipTrigger>
                <TooltipContent>
                  {t("pages.services.revision.detail.logs.tooltips.copy")}
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </div>
        <ScrollContainer logs={logs} />
        <div className="flex items-center justify-center pt-2 text-sm text-gray-11 dark:text-gray-dark-11">
          <span className="relative mr-2 flex h-3 w-3">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-gray-11 opacity-75 dark:bg-gray-dark-11"></span>
            <span className="relative inline-flex h-3 w-3 rounded-full bg-gray-11 dark:bg-gray-dark-11"></span>
          </span>
          {t("pages.services.revision.detail.logs.logsCount", {
            count: logs.length,
          })}
        </div>
        <Tabs
          value={selectedTab}
          className="flex w-full grid-rows-[auto_1fr] flex-col"
          onValueChange={(value) => {
            setSelectedTab(value);
          }}
        >
          <TabsList variant="boxed" className="w-max">
            {pods.map((pod, index) => (
              <TabsTrigger key={pod.name} variant="boxed" value={pod.name}>
                Pod {index}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      </Card>
    </div>
  );
};
