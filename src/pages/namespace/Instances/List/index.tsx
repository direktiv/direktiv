import { useInstances } from "~/api/instances/query/get";

const InstancesListPage = () => {
  const { data } = useInstances({ limit: 10, offset: 0 });

  return (
    <div>
      <h1>List</h1>
      {data?.instances.results.map((instance) => (
        <div key={instance.id}>
          {instance.as} <i>{instance.id}</i>
        </div>
      ))}
    </div>
  );
};

export default InstancesListPage;
