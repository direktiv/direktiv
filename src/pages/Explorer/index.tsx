import { FC } from "react";
import { useNamespaces } from "../../api/namespaces";

const ExplorerPage: FC = () => {
  const { data: namespaces } = useNamespaces();

  return (
    <div>
      <h1>Explorer</h1>
      <div className="font-bold">Namespaces</div>
      <div>
        {namespaces?.results.map((namespace) => (
          <div key={namespace.name}>{namespace.name}</div>
        ))}
      </div>
    </div>
  );
};

export default ExplorerPage;
