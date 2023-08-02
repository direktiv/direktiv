import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/design/Tabs";

import Button from "~/design/Button";
import Input from "./Input";
import { Maximize2 } from "lucide-react";
import Output from "./Output";
import { t } from "i18next";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useInstanceId } from "../state/instanceContext";
import { useState } from "react";
import { z } from "zod";

const InputOutput = () => {
  const instanceId = useInstanceId();
  const { data } = useInstanceDetails({ instanceId });
  const tabs = ["input", "output"] as const;

  const instanceIsFinished = data?.instance.status !== "pending";
  const [activeTab, setActiveTab] = useState<(typeof tabs)[number]>("input");

  console.log("🚀", instanceIsFinished);

  return (
    <div className="relative flex grow">
      <Button icon size="sm" variant="outline" className="absolute right-0">
        <Maximize2 />
      </Button>
      <Tabs
        value={activeTab}
        className="grid w-full grid-rows-[auto_1fr] border-2 border-yellow-500"
        onValueChange={(value) => {
          const tabValueParsed = z.enum(tabs).safeParse(value);
          if (tabValueParsed.success) {
            setActiveTab(tabValueParsed.data);
          }
        }}
      >
        <TabsList variant="boxed" className="w-max">
          <TabsTrigger variant="boxed" value={tabs[0]}>
            {t("pages.instances.detail.inputOutput.tabs.input")}
          </TabsTrigger>
          <TabsTrigger variant="boxed" value={tabs[1]}>
            {t("pages.instances.detail.inputOutput.tabs.output")}
          </TabsTrigger>
        </TabsList>
        <TabsContent
          value={tabs[0]}
          className="flex h-full grow border-2 border-green-500"
        >
          <div className="grow border">
            <Input />
          </div>
        </TabsContent>
        <TabsContent value={tabs[1]} className="grow">
          <Output instanceIsFinished={instanceIsFinished} />
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default InputOutput;
