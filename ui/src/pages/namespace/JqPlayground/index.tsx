import {
  BookOpen,
  FileInput,
  FileOutput,
  Play,
  PlaySquare,
} from "lucide-react";
import { Controller, SubmitHandler, useForm } from "react-hook-form";
import {
  ExecuteJqueryPayloadSchema,
  ExecuteJqueryPayloadType,
} from "~/api/jq/schema";
import { FC, useRef, useState } from "react";
import {
  useJqPlaygroundActions,
  useJqPlaygroundInput,
  useJqPlaygroundQuery,
} from "~/util/store/jqPlaygrpund";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import Editor from "~/design/Editor";
import Examples from "./Examples";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { decode } from "js-base64";
import { prettifyJsonString } from "~/util/helpers";
import { useExecuteJQuery } from "~/api/jq/mutate/executeQuery";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const defaultJx = "jq(.)";
const defaultData = "{}";

const JqPlaygroundPage: FC = () => {
  const { t } = useTranslation();
  const theme = useTheme();
  const {
    setInput: storeInputInLocalstorage,
    setQuery: storeQueryInLocalstorage,
  } = useJqPlaygroundActions();
  const jxFromStore = useJqPlaygroundQuery() ?? defaultJx;
  const dataFromStore = useJqPlaygroundInput() ?? defaultData;
  const formRef = useRef<HTMLFormElement>(null);

  const [output, setOutput] = useState("");

  const {
    register,
    handleSubmit,
    control,
    watch,
    setError,
    clearErrors,
    setValue,
    formState: { errors },
  } = useForm<ExecuteJqueryPayloadType>({
    resolver: zodResolver(ExecuteJqueryPayloadSchema),
    defaultValues: {
      data: dataFromStore,
      jx: jxFromStore,
    },
  });

  const { mutate: executeQuery, isPending } = useExecuteJQuery({
    onSuccess: (data) => {
      setOutput("");
      if (data.data.output[0]) {
        setOutput(decode(data.data.output[0]));
      }
    },
    onError: (error) => {
      setOutput("");
      setError("root", {
        message: error,
      });
    },
  });

  const onSubmit: SubmitHandler<ExecuteJqueryPayloadType> = (params) => {
    setOutput("");
    executeQuery(params);
  };

  const onRunSnippet = (params: ExecuteJqueryPayloadType) => {
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
                storeQueryInLocalstorage(e.target.value);
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
                        storeInputInLocalstorage(newData);
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
        </form>
      </Card>
      <Examples onRunSnippet={onRunSnippet} />
    </div>
  );
};

export default JqPlaygroundPage;
