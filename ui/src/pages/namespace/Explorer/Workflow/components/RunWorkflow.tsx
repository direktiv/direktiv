import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { Play } from "lucide-react";
import { useCreateInstance } from "~/api/instances/mutate/create";
import { useForm } from "react-hook-form";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";
import { workflowInputSchema } from "./utils";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  payload: string;
};

const defaultEmptyJson = "{\n    \n}";

const RunWorkflow = ({ path }: { path: string }) => {
  const { toast } = useToast();
  const { t } = useTranslation();
  const theme = useTheme();
  const navigate = useNavigate();

  const [jsonInput, setJsonInput] = useState(defaultEmptyJson);

  const { setValue } = useForm<FormInput>({
    defaultValues: {
      payload: defaultEmptyJson,
    },
    resolver: zodResolver(z.object({ payload: workflowInputSchema })),
  });

  const { mutate: runWorkflow, isPending } = useCreateInstance({
    onSuccess: (namespace, data) => {
      navigate({
        to: "/n/$namespace/instances/$id",
        params: { namespace, id: data.data.id },
      });
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
    runWorkflow({
      path,
      payload: jsonInput,
    });
  };

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
        <Card
          className="h-96 w-full p-4 sm:h-[500px]"
          noShadow
          background="weight-1"
          data-testid="run-workflow-editor"
        >
          <Editor
            value={jsonInput}
            onMount={(editor) => {
              editor.focus();
              if (jsonInput === defaultEmptyJson) {
                editor.setPosition({ lineNumber: 2, column: 5 });
              }
            }}
            onChange={(newData) => {
              if (newData != undefined) setJsonInput(newData);

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
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost" data-testid="run-workflow-cancel-btn">
            {t("pages.explorer.tree.workflow.runWorkflow.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          loading={isPending}
          onClick={runButtonOnClick}
          data-testid="run-workflow-submit-btn"
        >
          {!isPending && <Play />}
          {t("pages.explorer.tree.workflow.runWorkflow.runBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default RunWorkflow;
