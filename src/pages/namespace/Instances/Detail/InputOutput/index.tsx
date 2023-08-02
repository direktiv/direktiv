import Output from "./Output";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useInstanceId } from "../state/instanceContext";

const InputOutput = () => {
  const instanceId = useInstanceId();
  const { data } = useInstanceDetails({ instanceId });

  const instanceIsFinished = data?.instance.status !== "pending";

  return (
    <div className="grow">
      <Output instanceIsFinished={instanceIsFinished} />
    </div>
  );
};

export default InputOutput;
