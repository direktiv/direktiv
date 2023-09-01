import { useGroups } from "~/api/enterprise/groups/query/get";

const GroupsPage = () => {
  const { data } = useGroups();

  return (
    <div>
      <h1>groups</h1>
      {data?.groups.map((group) => (
        <div key={group.id}>{group.group}</div>
      ))}
    </div>
  );
};

export default GroupsPage;
