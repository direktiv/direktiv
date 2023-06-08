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
import { Play } from "lucide-react";
import { ScrollArea } from "~/design/ScrollArea";
import { useCreateTag } from "~/api/tree/mutate/createTag";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  payload: string;
};

const RunWorkflow = ({ path, close }: { path: string; close: () => void }) => {
  const { t } = useTranslation();
  const theme = useTheme();
  const {
    handleSubmit,
    formState: { isDirty, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver: zodResolver(z.object({})),
  });

  const [workflowInput, setWorkflowInput] = useState("{\n    \n}");

  // TODO: replace useCreateTag with useRunWorkflow
  const { mutate: runWorkflow, isLoading } = useCreateTag({
    onSuccess: () => {
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = () => {
    console.log("🚀 submitted");
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

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
          <Tabs defaultValue="json">
            <TabsList variant="boxed">
              <TabsTrigger variant="boxed" value="json">
                {t("pages.explorer.tree.workflow.runWorkflow.jsonInput")}
              </TabsTrigger>
              <TabsTrigger variant="boxed" value="formInput">
                {t("pages.explorer.tree.workflow.runWorkflow.formInput")}
              </TabsTrigger>
            </TabsList>
            <TabsContent value="json" asChild>
              <Card className="h-96 w-full p-4 sm:h-[500px]" noShadow>
                <Editor
                  value={workflowInput}
                  onMount={(editor) => {
                    editor.focus();
                    editor.setPosition({ lineNumber: 2, column: 5 });
                  }}
                  onChange={(newData) => {
                    if (newData) {
                      setWorkflowInput(newData);
                    }
                  }}
                  language="json"
                  theme={theme ?? undefined}
                />
              </Card>
            </TabsContent>
            <TabsContent value="formInput" asChild>
              <ScrollArea className="h-96 w-full p-4 sm:h-[500px]"></ScrollArea>
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
