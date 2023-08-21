import {
  PodLogsSubscriber,
  usePodLogs,
} from "~/api/services/query/revision/pods/getLogs";
import {
  PodsSubscriber,
  usePods,
} from "~/api/services/query/revision/pods/getAll";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";

import { Card } from "~/design/Card";
import { PodsListSchemaType } from "~/api/services/schema";
import { useState } from "react";

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
  const [selectedTab, setSelectedTab] = useState(pods[0]?.name ?? "");

  const { data: logData } = usePodLogs({
    name: selectedTab,
  });

  const pod = pods.find((pod) => pod.name === selectedTab);

  if (!pod) return null;

  return (
    <div>
      <PodLogsSubscriber name={selectedTab} />
      <Card className="p-5">
        <h1 className="font-bold">
          {pod.name.split("-").at(-1)} {pod.status}
        </h1>
        {logData?.data}
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
