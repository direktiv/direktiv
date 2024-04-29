import {
  BookOpen,
  FileInput,
  FileOutput,
  Play,
  PlaySquare,
} from "lucide-react";
import { FC, useState } from "react";
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
import { prettifyJsonString } from "~/util/helpers";
import { useExecuteJQuery } from "~/api/jq/mutate/executeQuery";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const JqPlaygroundPage: FC = () => {
  const { t } = useTranslation();
  const theme = useTheme();
  const {
    setInput: storeInputInLocalstorage,
    setQuery: storeQueryInLocalstorage,
  } = useJqPlaygroundActions();

  const [query, setQuery] = useState(useJqPlaygroundQuery() ?? ".");
  const [input, setInput] = useState(useJqPlaygroundInput() ?? "{}");
  const [output, setOutput] = useState("");
  const [error, setError] = useState("");

  const { mutate: executeQuery, isPending } = useExecuteJQuery({
    onSuccess: (data) => {
      setOutput(prettifyJsonString(data.results?.[0] ?? "{}"));
    },
    onError: (error) => {
      setOutput("");
      if (error) {
        setError(error);
      }
    },
  });

  const submitQuery = ({ query, input }: { query: string; input: string }) => {
    /**
     * Always clear the output before submiting a new query to the backend.
     * Otherwise when the request takes longer or produces an error the input
     * and output displayed to the user would not match.
     */
    setOutput("");
    executeQuery({ query, inputJsonString: input });
  };

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    submitQuery({ query, input });
  };

  const changeQuery = (newQuery: string) => {
    setQuery(newQuery);
    storeQueryInLocalstorage(newQuery);
    setError("");
  };

  const updateInput = (newData: string | undefined) => {
    if (newData === undefined) return;
    setInput(newData);
    storeInputInLocalstorage(newData);
    setError("");
  };

  const onRunSnippet = ({ query, input }: { query: string; input: string }) => {
    window.scrollTo({ top: 0, behavior: "smooth" });
    updateInput(prettifyJsonString(input));
    changeQuery(query);
    submitQuery({ query, input });
  };

  const formId = "jq-playground-form";
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
          id={formId}
          onSubmit={handleSubmit}
          className="flex flex-col gap-5"
        >
          <div className="flex flex-col gap-5 sm:flex-row">
            <Input
              data-testid="jq-query-input"
              placeholder={t("pages.jqPlayground.queryPlaceholder")}
              value={query}
              onChange={(e) => changeQuery(e.target.value)}
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
          {error && <FormErrors errors={{ error: { message: error } }} />}
          <div className="flex flex-col gap-5 md:flex-row">
            <Card className="flex h-96 w-full flex-col p-4" noShadow>
              <div className="mb-5 flex">
                <h3 className="flex grow items-center gap-x-2 font-medium">
                  <FileInput className="h-5" />
                  {t("pages.jqPlayground.input")}
                </h3>
                <CopyButton
                  value={input}
                  buttonProps={{
                    variant: "outline",
                    size: "sm",
                    type: "button",
                    disabled: !input,
                    "data-testid": "copy-input-btn",
                  }}
                />
              </div>
              <div data-testid="jq-input-editor" className="flex grow">
                <Editor
                  value={input}
                  language="json"
                  onChange={updateInput}
                  theme={theme ?? undefined}
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
