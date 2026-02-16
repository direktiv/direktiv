import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/design/Tabs";
import { useEffect, useState } from "react";

import Input from "./Input";
import Output from "./Output";
import { t } from "i18next";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useInstanceId } from "../../store/instanceContext";
import { z } from "zod";

const InputOutput = () => {
  const instanceId = useInstanceId();
  const { data } = useInstanceDetails({ instanceId });
  const tabs = ["input", "output"] as const;

  const [activeTab, setActiveTab] = useState<(typeof tabs)[number]>(
    data?.status === "complete" ? "output" : "input"
  );

  const outputIsPending =
    !!data && (data.status === "pending" || data.status === "running");

  useEffect(() => {
    setActiveTab(outputIsPending ? "input" : "output");
  }, [outputIsPending]);

  if (!data) return null;

  return (
    <div className="flex grow">
      <Tabs
        data-testid="inputOutputPanel"
        value={activeTab}
        className="flex w-full grid-rows-[auto_1fr] flex-col"
        onValueChange={(value) => {
          const tabValueParsed = z.enum(tabs).safeParse(value);
          if (tabValueParsed.success) {
            setActiveTab(tabValueParsed.data);
          }
        }}
      >
        <TabsContent value={tabs[0]} className="flex grow" asChild>
          <Input />
        </TabsContent>
        <TabsContent value={tabs[1]} className="flex grow" asChild>
          <Output isPending={outputIsPending} />
        </TabsContent>
        <TabsList variant="boxed" className="w-max">
          <TabsTrigger variant="boxed" value={tabs[0]}>
            {t("pages.instances.detail.inputOutput.tabs.input")}
          </TabsTrigger>
          <TabsTrigger variant="boxed" value={tabs[1]}>
            {t("pages.instances.detail.inputOutput.tabs.output")}
          </TabsTrigger>
        </TabsList>
      </Tabs>
    </div>
  );
};

export default InputOutput;
