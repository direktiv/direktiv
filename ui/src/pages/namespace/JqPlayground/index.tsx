import {
  BookOpen,
  FileInput,
  FileOutput,
  Play,
  PlaySquare,
  ScrollText,
} from "lucide-react";
import { Controller, SubmitHandler, useForm } from "react-hook-form";
import {
  ExecuteJxQueryPayloadSchema,
  ExecuteJxQueryPayloadType,
} from "~/api/jq/schema";
import { FC, useRef, useState } from "react";
import {
  useJqPlaygroundActions,
  useJqPlaygroundData,
  useJqPlaygroundJx,
} from "~/util/store/jqPlayground";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import Editor from "~/design/Editor";
import Examples from "./Examples";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { decode } from "js-base64";
import { prettifyJsonString } from "~/util/helpers";
import { useExecuteJxQuery } from "~/api/jq/mutate/executeQuery";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const defaultJx = "jq(.)";
const defaultData = "{}";
const defaultLogs = "";

const JqPlaygroundPage: FC = () => {
  const { t } = useTranslation();
  const theme = useTheme();
  const {
    setData: storePlaygroundDataInLocalstorage,
    setJx: storeJxInLocalstorage,
  } = useJqPlaygroundActions();
  const jxFromStore = useJqPlaygroundJx() ?? defaultJx;
  const dataFromStore = useJqPlaygroundData() ?? defaultData;
  const formRef = useRef<HTMLFormElement>(null);

  const [output, setOutput] = useState("");
  const [logs, setLogs] = useState(defaultLogs);

  const {
    register,
    handleSubmit,
    control,
    watch,
    setError,
    clearErrors,
    setValue,
    formState: { errors },
  } = useForm<ExecuteJxQueryPayloadType>({
    resolver: zodResolver(ExecuteJxQueryPayloadSchema),
    defaultValues: {
      data: dataFromStore,
      jx: jxFromStore,
    },
  });

  const clearLogsAndOutput = () => {
    setOutput("");
    setLogs(defaultLogs);
  };

  const { mutate: executeQuery, isPending } = useExecuteJxQuery({
    onSuccess: ({ data }) => {
      clearLogsAndOutput();
      if (data.output[0]) setOutput(decode(data.output[0]));
      if (data.logs) setLogs(decode(data.logs));
    },
    onError: (error) => {
      clearLogsAndOutput();
      setError("root", {
        message: error,
      });
    },
  });

  const onSubmit: SubmitHandler<ExecuteJxQueryPayloadType> = (params) => {
    clearLogsAndOutput();
    executeQuery(params);
  };

  const onRunSnippet = (params: ExecuteJxQueryPayloadType) => {
    window.scrollTo({ top: 0, behavior: "smooth" });
    setValue("data", prettifyJsonString(params.data));
    setValue("jx", params.jx);
    formRef.current?.requestSubmit();
  };

  const currentData = watch("data");

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <div className="flex">
        <h3 className="flex grow items-center gap-x-2 font-bold">
          <PlaySquare className="h-5" />
          {t("pages.jqPlayground.title")}
        </h3>
        <Button variant="outline" asChild isAnchor>
          <a
            href="https://stedolan.github.io/jq/manual/"
            target="_blank"
            rel="noopener noreferrer"
          >
            <BookOpen />
            {t("pages.jqPlayground.openManualBtn")}
          </a>
        </Button>
      </div>
      <Card className="p-5 text-sm ">
        {t("pages.jqPlayground.description")}
      </Card>
      <Card className="p-5">
        <form
          ref={formRef}
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col gap-5"
        >
          <div className="flex flex-col gap-5 sm:flex-row">
            <Input
              data-testid="jq-query-input"
              placeholder={t("pages.jqPlayground.queryPlaceholder")}
              {...register("jx")}
              onChange={(e) => {
                clearErrors();
                register("jx").onChange(e);
                storeJxInLocalstorage(e.target.value);
              }}
            />
            <Button
              data-testid="jq-run-btn"
              className="grow sm:w-44"
              type="submit"
              variant="primary"
              disabled={isPending}
              loading={isPending}
            >
              {!isPending && <Play />}
              {t("pages.jqPlayground.submitBtn")}
            </Button>
          </div>
          <FormErrors errors={errors} className="mb-5" />
          <div className="flex flex-col gap-5 md:flex-row">
            <Card className="flex h-96 w-full flex-col p-4" noShadow>
              <div className="mb-5 flex">
                <h3 className="flex grow items-center gap-x-2 font-medium">
                  <FileInput className="h-5" />
                  {t("pages.jqPlayground.input")}
                </h3>
                <CopyButton
                  value={currentData}
                  buttonProps={{
                    variant: "outline",
                    size: "sm",
                    type: "button",
                    disabled: !currentData,
                    "data-testid": "copy-input-btn",
                  }}
                />
              </div>
              <div data-testid="jq-input-editor" className="flex grow">
                <Controller
                  control={control}
                  name="data"
                  render={({ field }) => (
                    <Editor
                      value={field.value}
                      language="json"
                      onChange={(newData) => {
                        if (newData === undefined) return;
                        clearErrors();
                        field.onChange(newData);
                        storePlaygroundDataInLocalstorage(newData);
                      }}
                      theme={theme ?? undefined}
                    />
                  )}
                />
              </div>
            </Card>
            <Card className="flex h-96 w-full flex-col p-4" noShadow>
              <div className="mb-5 flex">
                <h3 className="flex grow items-center gap-x-2 font-medium">
                  <FileOutput className="h-5" />
                  {t("pages.jqPlayground.output")}
                </h3>
                <CopyButton
                  value={output}
                  buttonProps={{
                    variant: "outline",
                    size: "sm",
                    type: "button",
                    disabled: !output,
                    "data-testid": "copy-output-btn",
                  }}
                />
              </div>
              <div data-testid="jq-output-editor" className="flex grow">
                <Editor
                  language="json"
                  value={output}
                  options={{
                    readOnly: true,
                  }}
                  theme={theme ?? undefined}
                />
              </div>
            </Card>
          </div>

          <Card className="flex h-32 w-full flex-col p-4" noShadow>
            <div className="mb-5 flex">
              <h3 className="flex grow items-center gap-x-2 font-medium">
                <ScrollText className="h-5" />
                {t("pages.jqPlayground.logs")}
              </h3>
              <CopyButton
                value={logs}
                buttonProps={{
                  variant: "outline",
                  size: "sm",
                  type: "button",
                  disabled: !logs,
                  "data-testid": "copy-logs-btn",
                }}
              />
            </div>
            <div data-testid="jq-logs-editor" className="flex grow">
              <Editor
                language="shell"
                value={logs}
                options={{
                  readOnly: true,
                }}
                theme={theme ?? undefined}
              />
            </div>
          </Card>
        </form>
      </Card>
      <Examples onRunSnippet={onRunSnippet} />
    </div>
  );
};

export default JqPlaygroundPage;
