import { FC, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import Input from "~/design/Input";
import { Play } from "lucide-react";
import { useExecuteJQuery } from "~/api/jq/mutate/executeQuery";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const data = {
  foo: [
    { name: "JSON", good: true },
    { name: "XML", good: false },
  ],
};

const JqPlaygroundPage: FC = () => {
  const { t } = useTranslation();
  const theme = useTheme();
  const [result, setResult] = useState("");
  const { mutate: executeQuery, isLoading } = useExecuteJQuery({
    onSuccess: (data) => {
      setResult(JSON.stringify(data.results, null, 2));
    },
  });
  const [query, setQuery] = useState(".foo[1]"); // TODO: remove default query
  const [input, setInput] = useState(JSON.stringify(data, null, 2)); // TODO: remove default query

  const formId = "jq-playground-form";

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    executeQuery({ query, inputJsonString: input });
  };

  return (
    <Card className="m-5 p-5">
      <form id={formId} onSubmit={handleSubmit} className="flex flex-col gap-5">
        <div className="flex flex-col gap-5 sm:flex-row">
          <Input
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
            }}
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
        <div className="flex gap-5">
          <Card className="h-96 w-full p-4" noShadow background="weight-1">
            <Editor
              value={input}
              language="json"
              onChange={(newData) => {
                if (newData) {
                  setInput(newData);
                }
              }}
              theme={theme ?? undefined}
            />
          </Card>
          <Card className="h-96 w-full p-4" noShadow background="weight-1">
            <Editor
              language="json"
              value={result}
              options={{
                readOnly: true,
              }}
              theme={theme ?? undefined}
            />
          </Card>
        </div>
      </form>
    </Card>
  );
};

export default JqPlaygroundPage;
