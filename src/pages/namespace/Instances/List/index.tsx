import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
import { useInstances } from "~/api/instances/query/get";

const InstancesListPage = () => {
  const { data } = useInstances({ limit: 10, offset: 0 });

  return (
    <div>
      <h1>List</h1>
      {data?.instances.results.map((instance) => (
        <Link
          to={pages.instances.createHref({
            namespace: data.namespace,
            instance: instance.id,
          })}
          key={instance.id}
        >
          {instance.as} <i>{instance.id}</i>
        </Link>
      ))}
    </div>
  );
};

export default InstancesListPage;
