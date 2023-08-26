import { FC, useState } from "react";

import Button from "~/design/Button";
import Input from "~/design/Input";
import { Play } from "lucide-react";
import { useExecuteJQuery } from "~/api/jq/mutate/executeQuery";
import { useTranslation } from "react-i18next";

const JqPlaygroundPage: FC = () => {
  const { t } = useTranslation();
  const { mutate: executeQuery } = useExecuteJQuery();
  const [query, setQuery] = useState(".foo[1]");

  const data = {
    foo: [
      { name: "JSON", good: true },
      { name: "XML", good: false },
    ],
  };

  return (
    <div className="flex flex-col space-y-10 p-5">
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
    </div>
  );
};

export default JqPlaygroundPage;
