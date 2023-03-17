import { FolderOpen, FolderUp, Play } from "lucide-react";
import { useNamespace, useNamespaceActions } from "../../util/store/namespace";

import Button from "../../componentsNext/Button";
import { FC } from "react";
import { Link } from "react-router-dom";
import moment from "moment";
import { pages } from "../../util/router/pages";
import { useNamespaces } from "../../api/namespaces";
import { useTree } from "../../api/tree";

const ExplorerPage: FC = () => {
  const { data: namespaces } = useNamespaces();
  const selectedNamespace = useNamespace();
  const { setNamespace } = useNamespaceActions();
  const { directory } = pages.explorer.useParams();

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
                <span className="flex-1">{file.name}</span>
                <span className="text-gray-gray8 dark:text-grayDark-gray8">
                  {moment(file.updatedAt).fromNow()}
                </span>
              </Link>
            )}

            {file.type === "workflow" && (
              <div className="flex items-center space-x-3">
                <Play />
                <span className="flex-1">{file.name}</span>
                <span className="text-gray-gray8 dark:text-grayDark-gray8">
                  {moment(file.updatedAt).fromNow()}
                </span>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

export default ExplorerPage;
