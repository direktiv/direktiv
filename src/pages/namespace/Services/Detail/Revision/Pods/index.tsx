import { Box, ScrollText } from "lucide-react";
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
import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import { NoResult } from "~/design/Table";
import { PodsListSchemaType } from "~/api/services/schema/pods";
import ScrollContainer from "./ScrollContainer";
import { podStatusToBadgeVariant } from "../../../components/utils";
import { twMergeClsx } from "~/util/helpers";
import { useState } from "react";
import { useTranslation } from "react-i18next";

export const Pods = ({
  service,
  revision,
  workflow,
  version,
  className,
}: {
  service: string;
  revision: string;
  workflow?: string;
  version?: string;
  className?: string;
}) => {
  const { data: podsList, isSuccess } = usePods({
    revision,
    service,
    workflow,
    version,
  });
  if (!isSuccess) return null;

  return (
    <>
      <PodsSubscriber
        revision={revision}
        service={service}
        workflow={workflow}
        version={version}
      />
      <PodsWithData pods={podsList.pods} className={className} />
    </>
  );
};

export const PodsWithData = ({
  pods,
  className,
}: {
  pods: PodsListSchemaType["pods"];
  className?: string;
}) => {
  const { t } = useTranslation();
  const [selectedTab, setSelectedTab] = useState(pods[0]?.name ?? "");
  const { data: logData } = usePodLogs({
    name: selectedTab,
  });

  const pod = pods.find((pod) => pod.name === selectedTab);

  const logs = logData?.data.split("\n") ?? [];

  if (!pod)
    return (
      <Card className="m-5 flex grow">
        <NoResult icon={Box}>
          {t("pages.services.revision.detail.logs.noPods")}
        </NoResult>
      </Card>
    );

  return (
    <div className="grid grow p-5">
      <PodLogsSubscriber name={selectedTab} />
      <Card className="grid grid-rows-[auto_1fr_auto] p-5">
        <div className="mb-5 flex flex-col gap-5 sm:flex-row">
          <h3 className="flex grow items-center gap-x-2 font-medium">
            <ScrollText className="h-5" />
            {t("pages.services.revision.detail.logs.title", {
              name: pod.name.split("-").at(-1),
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
        <div
          className={twMergeClsx(
            "relative grid grow grid-rows-[1fr_auto]",
            "h-[calc(100vh-1rem)]",
            "md:h-[calc(100vh-30rem)]",
            "lg:h-[calc(100vh-26rem)]",
            className
          )}
        >
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
        </div>
        <Tabs
          value={selectedTab}
          onValueChange={(value) => {
            setSelectedTab(value);
          }}
        >
          <TabsList variant="boxed">
            {pods.map((pod, index, src) => (
              <TabsTrigger key={pod.name} variant="boxed" value={pod.name}>
                {t("pages.services.revision.detail.logs.tab", {
                  number: index + 1,
                  total: src.length,
                })}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      </Card>
    </div>
  );
};
