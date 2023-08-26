import Button from "~/design/Button";
import { FC } from "react";
import { useExecuteJQuery } from "~/api/jq/mutate/executeQuery";

const JqPlaygroundPage: FC = () => {
  const { mutate: executeQuery } = useExecuteJQuery();
  const query = ".foo[1]";
  const data = {
    foo: [
      { name: "JSON", good: true },
      { name: "XML", good: false },
    ],
  };

  return (
    <div className="flex flex-col space-y-10 p-5">
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
