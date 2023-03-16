import { FolderOpen, FolderUp, Play } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { useNamespace, useNamespaceActions } from "../../util/store/namespace";

import Button from "../../componentsNext/Button";
import { FC } from "react";
import { pages } from "../../util/router/pages";
import { useNamespaces } from "../../api/namespaces";
import { useTree } from "../../api/tree";

const ExplorerPage: FC = () => {
  const { data: namespaces } = useNamespaces();
  const selectedNamespace = useNamespace();
  const { setNamespace } = useNamespaceActions();

  const { directory } = useParams();
  const { data } = useTree({
    directory,
  });

  return (
    <div>
      <h1>Explorer {directory}</h1>
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
      <div className="flex flex-col space-y-5 ">
        {directory && (
          <Link
            to={pages.explorer.createHref({
              directory: directory.split("/").slice(0, -1).join("/"),
            })}
            className="flex items-center space-x-3"
          >
            <FolderUp />
            <span>..</span>
          </Link>
        )}
        {data?.children.results.map((file) => (
          <div key={file.name}>
            {file.type === "directory" && (
              <Link
                to={pages.explorer.createHref({
                  directory: directory
                    ? `${directory}/${file.name}`
                    : file.name,
                })}
                className="flex items-center space-x-3"
              >
                <FolderOpen />
                <span>{file.name}</span>
              </Link>
            )}

            {file.type === "workflow" && (
              <div className="flex items-center space-x-3">
                <Play />
                <span>{file.name}</span>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

export default ExplorerPage;
