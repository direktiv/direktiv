import { Card } from "~/design/Card";
import { NoPermissions } from "~/design/Table";
import PolicyEditor from "./Editor";
import { usePolicy } from "~/api/enterprise/policy/query/get";

const PolicyPage = () => {
  const { data, isAllowed, noPermissionMessage } = usePolicy();

  return (
    <div className="flex grow flex-col space-y-4 p-5">
      {isAllowed ? (
        <>{data && <PolicyEditor policyFromServer={data.body} />}</>
      ) : (
        <Card className="flex grow flex-col p-4">
          <NoPermissions>{noPermissionMessage}</NoPermissions>
        </Card>
      )}
    </div>
  );
};

export default PolicyPage;
