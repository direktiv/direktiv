import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/design/Tabs";
import { getValidationSchemaFromYaml, workflowInputSchema } from "./utils";
import { useEffect, useRef, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import FormInputHint from "./FormInputHint";
import { JSONSchemaForm } from "~/design/JSONschemaForm";
import { Play } from "lucide-react";
import { ScrollArea } from "~/design/ScrollArea";
import { decode } from "js-base64";
import { pages } from "~/util/router/pages";
import { useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { useNode } from "~/api/filesTree/query/node";
import { useRunWorkflow } from "~/api/tree/mutate/runWorkflow";
import { useTheme } from "~/util/store/theme";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  payload: string;
};

type JSONSchemaFormSubmit = Parameters<typeof JSONSchemaForm>[0]["onSubmit"];

const RunWorkflow = ({ path }: { path: string }) => {
  const { toast } = useToast();
  const { t } = useTranslation();
  const theme = useTheme();
  const navigate = useNavigate();
  const { data } = useNode({ path });
  const submitButtonRef = useRef<HTMLButtonElement>(null);
  const validationSchema = getValidationSchemaFromYaml(
    decode(data?.file.data ?? "")
  );

  // tab handling
  const isFormAvailable = validationSchema !== null;
  const tabs = ["json", "form"] as const;
  const [activeTab, setActiveTab] = useState<(typeof tabs)[number]>(
    isFormAvailable ? "form" : "json"
  );

  // it is possible that no data (or stale cache data) is available when this component mounts
  // and the initial value of activeTab is out of synch with the actual isFormAvailable value
  useEffect(() => {
    setActiveTab(isFormAvailable ? "form" : "json");
  }, [isFormAvailable]);

  const {
    setValue,
    getValues,
    formState: { isValid },
  } = useForm<FormInput>({
    defaultValues: {
      payload: "{\n    \n}",
    },
    resolver: zodResolver(z.object({ payload: workflowInputSchema })),
  });

  const { mutate: runWorkflow, isLoading } = useRunWorkflow({
    onSuccess: ({ namespace, instance }) => {
      navigate(pages.instances.createHref({ namespace, instance }));
    },
    onError: (error) => {
      toast({
        title: t("api.generic.error"),
        description:
          error ??
          t("pages.explorer.tree.workflow.runWorkflow.genericRunError"),
        variant: "error",
      });
    },
  });

  const runButtonOnClick = () => {
    // if this workflow supports a JSON form and the json form
    // tab is active we need to trigger this form via a ref to an
    // invisble submit button (this should be optimized but a ref
    // to the form did not work)
    if (isFormAvailable && activeTab === "form") {
      // this will implcitly trigger the JSONschema forms onSubmit callback
      submitButtonRef.current?.click();
    }

    if (activeTab === "json") {
      runWorkflow({
        path,
        payload: getValues("payload"),
      });
    }
  };

  const jsonSchemaFormSubmit: JSONSchemaFormSubmit = (form) => {
    runWorkflow({ path, payload: JSON.stringify(form.formData) });
  };

  const disableSubmit = !isValid;

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Play /> {t("pages.explorer.tree.workflow.runWorkflow.title")}
        </DialogTitle>
      </DialogHeader>
      <div
        className="my-3 flex flex-col gap-y-5"
        data-testid="run-workflow-dialog"
      >
        <Tabs
          value={activeTab}
          onValueChange={(value) => {
            const tabValueParsed = z.enum(tabs).safeParse(value);
            if (tabValueParsed.success) {
              setActiveTab(tabValueParsed.data);
            }
          }}
        >
          <TabsList variant="boxed">
            <TabsTrigger
              variant="boxed"
              value={tabs[0]}
              data-testid="run-workflow-json-tab-btn"
            >
              {t("pages.explorer.tree.workflow.runWorkflow.jsonInput")}
            </TabsTrigger>
            <TabsTrigger
              variant="boxed"
              value={tabs[1]}
              data-testid="run-workflow-form-tab-btn"
            >
              {t("pages.explorer.tree.workflow.runWorkflow.formInput")}
            </TabsTrigger>
          </TabsList>
          <TabsContent value={tabs[0]} asChild>
            <Card
              className="h-96 w-full p-4 sm:h-[500px]"
              noShadow
              background="weight-1"
              data-testid="run-workflow-editor"
            >
              <Editor
                value={getValues("payload")}
                onMount={(editor) => {
                  editor.focus();
                  editor.setPosition({ lineNumber: 2, column: 5 });
                }}
                onChange={(newData) => {
                  if (typeof newData === "string") {
                    setValue("payload", newData, {
                      shouldValidate: true,
                    });
                  }
                }}
                language="json"
                theme={theme ?? undefined}
              />
            </Card>
          </TabsContent>
          <TabsContent value={tabs[1]} asChild>
            <Card className="h-96 w-full p-4 sm:h-[500px]">
              {isFormAvailable ? (
                <ScrollArea className="h-full">
                  <JSONSchemaForm
                    schema={validationSchema}
                    action="submit"
                    onSubmit={jsonSchemaFormSubmit}
                  >
                    <Button
                      type="submit"
                      ref={submitButtonRef}
                      className="hidden"
                    />
                  </JSONSchemaForm>
                </ScrollArea>
              ) : (
                <FormInputHint />
              )}
            </Card>
          </TabsContent>
        </Tabs>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost" data-testid="run-workflow-cancel-btn">
            {t("pages.explorer.tree.workflow.runWorkflow.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          onClick={runButtonOnClick}
          data-testid="run-workflow-submit-btn"
        >
          {!isLoading && <Play />}
          {t("pages.explorer.tree.workflow.runWorkflow.runBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default RunWorkflow;
