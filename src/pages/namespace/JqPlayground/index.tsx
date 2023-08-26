import { FC, useState } from "react";

import Button from "~/design/Button";
import Input from "~/design/Input";
import { useExecuteJQuery } from "~/api/jq/mutate/executeQuery";

const JqPlaygroundPage: FC = () => {
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
      <Input
        value={query}
        onChange={(e) => {
          setQuery(e.target.value);
        }}
      />
      <Button
        onClick={() => {
          executeQuery({ query, inputJSON: JSON.stringify(data) });
        }}
      >
        Submit
      </Button>
    </div>
  );
};

export default JqPlaygroundPage;
