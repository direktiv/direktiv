import { FolderOpen, FolderUp, Github, Play } from "lucide-react";

import { FC } from "react";
import { Link } from "react-router-dom";
import moment from "moment";
import { pages } from "../../../util/router/pages";
import { useNamespace } from "../../../util/store/namespace";
import { useTree } from "../../../api/tree";

const ExplorerPage: FC = () => {
  const namespace = useNamespace();
  const { directory } = pages.explorer.useParams();

  const { data } = useTree({
    directory,
  });

  if (!namespace) return null;
  return (
    <div className="flex flex-col space-y-5 p-10">
      <div className="flex flex-col space-y-5 ">
        {directory && (
          <Link
            to={pages.explorer.createHref({
              namespace,
              directory: directory.split("/").slice(0, -1).join("/"),
            })}
            className="flex items-center space-x-3"
          >
            <FolderUp />
            <span>..</span>
          </Link>
        )}
        {data?.children?.results.map((file) => {
          let Icon = FolderOpen;
          if (file.expandedType === "workflow") {
            Icon = Play;
          }
          if (file.expandedType === "git") {
            Icon = Github;
          }

          const linkTarget =
            file.expandedType === "workflow"
              ? pages.workflow.createHref({
                  namespace,
                  file: directory ? `${directory}/${file.name}` : file.name,
                })
              : pages.explorer.createHref({
                  namespace,
                  directory: directory
                    ? `${directory}/${file.name}`
                    : file.name,
                });

          return (
            <div key={file.name}>
              <Link to={linkTarget} className="flex items-center space-x-3">
                <Icon />
                <span className="flex-1">{file.name}</span>
                <span className="text-gray-gray8 dark:text-grayDark-gray8">
                  {moment(file.updatedAt).fromNow()}
                </span>
              </Link>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default ExplorerPage;
