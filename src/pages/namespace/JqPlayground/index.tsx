import { FC, useState } from "react";
import { FileInput, FileOutput, Play, PlaySquare } from "lucide-react";
import {
  useJqPlaygroundActions,
  useJqPlaygroundInput,
  useJqPlaygroundQuery,
} from "~/util/store/jqPlaygrpund";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import Editor from "~/design/Editor";
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import { JqQueryErrorSchema } from "~/api/jq/schema";
import cheatsheet from "./cheatsheet";
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
  const [result, setResult] = useState("");
  const [error, setError] = useState("");

  const { mutate: executeQuery, isLoading } = useExecuteJQuery({
    onSuccess: (data) => {
      const resultAsJson = JSON.parse(data.results?.[0] ?? "{}");
      setResult(JSON.stringify(resultAsJson, null, 4));
    },
    onError: (error) => {
      setResult("");
      const errorParsed = JqQueryErrorSchema.safeParse(error);
      if (errorParsed.success) {
        setError(errorParsed.data.message);
      }
    },
  });

  const formId = "jq-playground-form";

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    executeQuery({ query, inputJsonString: input });
  };

  const onQueryChange = (newQuery: string) => {
    setQuery(newQuery);
    storeQueryInLocalstorage(newQuery);
    setError("");
  };

  const onInputChange = (newData: string | undefined) => {
    if (newData) {
      setInput(newData);
      storeInputInLocalstorage(newData);
      setError("");
    }
  };

  const onTemplateClick = ({
    query,
    input,
  }: {
    query: string;
    input: string;
  }) => {
    onInputChange(JSON.stringify(JSON.parse(input), null, 4));
    onQueryChange(query);
    executeQuery({ query, inputJsonString: input });
  };

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold">
        <PlaySquare className="h-5" />
        {t("pages.jqPlayground.title")}
      </h3>
      <Card className="p-5">
        <form
          id={formId}
          onSubmit={handleSubmit}
          className="flex flex-col gap-5"
        >
          <div className="flex flex-col gap-5 sm:flex-row">
            <Input
              placeholder={t("pages.jqPlayground.queryPlaceholder")}
              value={query}
              onChange={(e) => onQueryChange(e.target.value)}
            />
            <Button
              className="grow sm:w-64"
              type="submit"
              disabled={isLoading}
              loading={isLoading}
            >
              {!isLoading && <Play />}
              {t("pages.jqPlayground.submitBtn")}
            </Button>
          </div>
          {error && <FormErrors errors={{ error: { message: error } }} />}
          <div className="flex flex-col gap-5 md:flex-row">
            <Card className="flex h-96 w-full flex-col p-4" noShadow>
              <div className="mb-5 flex">
                <h3 className="flex grow items-center gap-x-2 font-medium">
                  <FileInput className="h-5" />
                  {t("pages.jqPlayground.output")}
                </h3>
                <CopyButton
                  value={input}
                  buttonProps={{
                    variant: "outline",
                    size: "sm",
                    type: "button",
                    disabled: !input,
                  }}
                />
              </div>
              <div className="flex grow">
                <Editor
                  value={input}
                  language="json"
                  onChange={onInputChange}
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
                  value={result}
                  buttonProps={{
                    variant: "outline",
                    size: "sm",
                    type: "button",
                    disabled: !result,
                  }}
                />
              </div>
              <div className="flex grow">
                <Editor
                  language="json"
                  value={result}
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
      <Card className="grid grid-cols-2 gap-5 p-5">
        {cheatsheet.map(({ query, input, tip, example }, index) => (
          <Card key={index} className="flex gap-2 p-5">
            <div className="grid grow grid-cols-2">
              <div className="font-mono text-primary-500">{example}</div>
              <div>{tip}</div>
            </div>
            <Button
              size="sm"
              variant="outline"
              onClick={() =>
                onTemplateClick({
                  query,
                  input,
                })
              }
            >
              run
            </Button>
          </Card>
        ))}
      </Card>
    </div>
  );
};

export default JqPlaygroundPage;
