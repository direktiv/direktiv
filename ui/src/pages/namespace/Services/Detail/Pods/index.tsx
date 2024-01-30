import { Box, ScrollText } from "lucide-react";
import { PodLogsSubscriber, usePodLogs } from "~/api/services/query/podLogs";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import { NoResult } from "~/design/Table";
import { PodSchemaType } from "~/api/services/schema/pods";
import ScrollContainer from "./ScrollContainer";
import { twMergeClsx } from "~/util/helpers";
import { usePods } from "~/api/services/query/pods";
import { useState } from "react";
import { useTranslation } from "react-i18next";

export const Pods = ({
  serviceId,
  className,
}: {
  serviceId: string;
  className?: string;
}) => {
  const { data: podsList, isLoading } = usePods(serviceId);

  if (isLoading) return null;

  return (
    <PodsWithData
      pods={podsList?.data ?? []}
      serviceId={serviceId}
      className={className}
    />
  );
};

const PodsWithData = ({
  pods,
  serviceId,
  className,
}: {
  serviceId: string;
  pods: PodSchemaType[];
  className?: string;
}) => {
  const { t } = useTranslation();
  const [selectedPod, setSelectedPod] = useState(pods?.[0]?.id ?? "");
  const { data: logData } = usePodLogs({
    pod: selectedPod,
    service: serviceId,
  });

  const logs = logData?.trim().split("\n") ?? [];
  const pod = pods.find((pod) => pod.id === selectedPod);

  if (!pod)
    return (
      <Card className="m-5 flex grow">
        <NoResult icon={Box}>{t("pages.services.detail.logs.noPods")}</NoResult>
      </Card>
    );

  return (
    <div className="grid grow p-5">
      <PodLogsSubscriber service={serviceId} pod={selectedPod} />
      <Card className="grid grid-rows-[auto_1fr_auto] p-5">
        <div className="mb-5 flex flex-col gap-5 sm:flex-row">
          <h3 className="flex grow items-center gap-x-2 font-medium">
            <ScrollText className="h-5" />
            {t("pages.services.detail.logs.title", {
              name: pod.id,
            })}
          </h3>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <CopyButton
                  value={logData ?? ""}
                  buttonProps={{
                    variant: "outline",
                    size: "sm",
                    disabled: !logData,
                  }}
                />
              </TooltipTrigger>
              <TooltipContent>
                {t("pages.services.detail.logs.tooltips.copy")}
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
            {t("pages.services.detail.logs.logsCount", {
              count: logs.length,
            })}
          </div>
        </div>
        <Tabs
          value={selectedPod}
          onValueChange={(value) => {
            setSelectedPod(value);
          }}
        >
          <TabsList variant="boxed">
            {pods.map((pod, index, src) => (
              <TabsTrigger key={pod.id} variant="boxed" value={pod.id}>
                {t("pages.services.detail.logs.tab", {
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
