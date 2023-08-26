import { FC, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import Input from "~/design/Input";
import { Play } from "lucide-react";
import { set } from "date-fns";
import { useExecuteJQuery } from "~/api/jq/mutate/executeQuery";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const JqPlaygroundPage: FC = () => {
  const { t } = useTranslation();
  const theme = useTheme();
  const [result, setResult] = useState("");
  const { mutate: executeQuery } = useExecuteJQuery({
    onSuccess: (data) => {
      setResult(JSON.stringify(data.results));
    },
  });
  const [query, setQuery] = useState(".foo[1]");

  const data = {
    foo: [
      { name: "JSON", good: true },
      { name: "XML", good: false },
    ],
  };

  return (
    <Card className="m-5 flex flex-col gap-5 p-5">
      <div className="flex flex-col gap-5 sm:flex-row">
        <Input
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
          }}
        />
        <Button
          className="grow sm:w-64"
          onClick={() => {
            executeQuery({ query, inputJSON: JSON.stringify(data) });
          }}
        >
          <Play />
          {t("pages.jqPlayground.submitBtn")}
        </Button>
      </div>
      <div className="flex gap-5">
        <Card className="h-96 w-full p-4" noShadow background="weight-1">
          <Editor
            value={JSON.stringify(data, null, 2)}
            language="json"
            onChange={(newData) => {
              // if (newData) {
              //   setWorkflowData(newData);
              //   setValue("fileContent", newData);
              // }
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
    </Card>
  );
};

export default JqPlaygroundPage;
