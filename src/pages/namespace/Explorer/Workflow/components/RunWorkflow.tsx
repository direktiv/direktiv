import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { SubmitHandler, useForm } from "react-hook-form";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/design/Tabs";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import FormInputHint from "./FormInputHint";
import { JSONSchemaForm } from "~/design/JSONschemaForm";
import { Play } from "lucide-react";
import { ScrollArea } from "~/design/ScrollArea";
import { getValidationSchema } from "./utils";
import { pages } from "~/util/router/pages";
import { useNavigate } from "react-router-dom";
import { useNodeContent } from "~/api/tree/query/node";
import { useRunWorkflow } from "~/api/tree/mutate/runWorkflow";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  payload: string;
};

const RunWorkflow = ({ path }: { path: string }) => {
  const { t } = useTranslation();
  const theme = useTheme();
  const navigate = useNavigate();
  const { data } = useNodeContent({ path });
  const validationSchema = getValidationSchema(
    data?.revision?.source && atob(data?.revision?.source)
  );

  const formAvailable = validationSchema !== null;
  const tabs = ["json", "form"] as const;
  const activeTab: (typeof tabs)[number] = formAvailable ? "form" : "json";

  const {
    handleSubmit,
    setValue,
    getValues,
    formState: { isValid },
  } = useForm<FormInput>({
    defaultValues: {
      payload: "{\n    \n}",
    },
    resolver: zodResolver(
      z.object({
        payload: z.string().refine((string) => {
          try {
            JSON.parse(string);
            return true;
          } catch (error) {
            return false;
          }
        }),
      })
    ),
  });

  const { mutate: runWorkflow, isLoading } = useRunWorkflow({
    onSuccess: ({ namespace, instance }) => {
      navigate(pages.instances.createHref({ namespace, instance }));
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({ payload }) => {
    runWorkflow({
      path,
      payload,
    });
  };

  const disableSubmit = !isValid;

  const formId = `run-workflow-${path}`;
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Play /> {t("pages.explorer.tree.workflow.runWorkflow.title")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3 flex flex-col gap-y-5">
        <form id={formId} onSubmit={handleSubmit(onSubmit)}>
          <Tabs defaultValue={activeTab}>
            <TabsList variant="boxed">
              <TabsTrigger variant="boxed" value={tabs[0]}>
                {t("pages.explorer.tree.workflow.runWorkflow.jsonInput")}
              </TabsTrigger>
              <TabsTrigger variant="boxed" value={tabs[1]}>
                {t("pages.explorer.tree.workflow.runWorkflow.formInput")}
              </TabsTrigger>
            </TabsList>
            <TabsContent value={tabs[0]} asChild>
              <Card className="h-96 w-full p-4 sm:h-[500px]" noShadow>
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
              {formAvailable ? (
                <ScrollArea className="h-96 w-full p-4 sm:h-[500px]">
                  <JSONSchemaForm schema={validationSchema}></JSONSchemaForm>
                </ScrollArea>
              ) : (
                <FormInputHint />
              )}
            </TabsContent>
          </Tabs>
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.workflow.runWorkflow.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          form={formId}
          data-testid="dialog-create-tag-btn-submit"
        >
          {!isLoading && <Play />}
          {t("pages.explorer.tree.workflow.runWorkflow.runBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default RunWorkflow;
