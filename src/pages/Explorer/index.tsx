import { useNamespace, useNamespaceActions } from "../../util/store/namespace";

import Button from "../../componentsNext/Button";
import { FC } from "react";
import { useNamespaces } from "../../api/namespaces";

const ExplorerPage: FC = () => {
  const { data: namespaces } = useNamespaces();
  const selectedNamespace = useNamespace();
  const { setNamespace } = useNamespaceActions();

  return (
    <div>
      <h1>Explorer</h1>
      <div className="py-5 font-bold">Namespaces</div>
      <div className="flex flex-col space-y-5 ">
        {namespaces?.results.map((namespace) => (
          <Button
            key={namespace.name}
            onClick={() => {
              setNamespace(namespace.name);
            }}
            color="primary"
            size="sm"
          >
            {selectedNamespace === namespace.name && "✅"} {namespace.name}
          </Button>
        ))}
      </div>
      <div className="py-5 font-bold">Files</div>
      <div className="flex flex-col space-y-5 "></div>
    </div>
  );
};

export default ExplorerPage;
