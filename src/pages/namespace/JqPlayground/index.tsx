import Button from "~/design/Button";
import { FC } from "react";
import { useExecuteJQuery } from "~/api/jq/mutate/executeQuery";

const JqPlaygroundPage: FC = () => {
  const { mutate: executeQuery } = useExecuteJQuery();
  const query = ".foo[1]";
  const data =
    "eyJmb28iOiBbeyJuYW1lIjoiSlNPTiIsICJnb29kIjp0cnVlfSwgeyJuYW1lIjoiWE1MIiwgImdvb2QiOmZhbHNlfV19";

  return (
    <div className="flex flex-col space-y-10 p-5">
      <Button
        onClick={() => {
          executeQuery({ query, inputJSON: data });
        }}
      >
        Submit
      </Button>
    </div>
  );
};

export default JqPlaygroundPage;
