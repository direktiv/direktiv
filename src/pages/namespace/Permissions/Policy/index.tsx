import PolicyEditor from "./Editor";
import { usePolicy } from "~/api/enterprise/policy/query/get";

const PolicyPage = () => {
  const { data } = usePolicy();
  return (
    <div className="flex grow flex-col space-y-4 p-5">
      {data && <PolicyEditor policyFromServer={data} />}
    </div>
  );
};

export default PolicyPage;
