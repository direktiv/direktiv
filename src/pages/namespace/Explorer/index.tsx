import { FolderOpen, FolderTree, FolderUp, Github, Play } from "lucide-react";

import Button from "../../../componentsNext/Button";
import { FC } from "react";
import { Link } from "react-router-dom";
import { RxChevronDown } from "react-icons/rx";
import { analyzePath } from "../../../util/router/utils";
import moment from "moment";
import { pages } from "../../../util/router/pages";
import { useNamespace } from "../../../util/store/namespace";
import { useTree } from "../../../api/tree";

const ExplorerPage: FC = () => {
  const namespace = useNamespace();
  const { path } = pages.explorer.useParams();

  const { data } = useTree({ path });
  const { parent, isRoot } = analyzePath(path);

  if (!namespace) return null;

  return (
    <div className="flex flex-col space-y-5 p-5 text-sm">
      <div className="flex flex-col text-base max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between">
        <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
          <FolderTree className="h-5" />
          {data?.node?.path}
        </h3>
        <Button variant="primary">
          Actions <RxChevronDown />
        </Button>
      </div>
      <div className="flex flex-col space-y-5 ">
        {!isRoot && (
          <Link
            to={pages.explorer.createHref({
              namespace,
              path: parent?.absolute,
            })}
            className="flex items-center space-x-3"
          >
            <FolderUp className="h-5" />
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
                  path: file.path,
                })
              : pages.explorer.createHref({
                  namespace,
                  path: file.path,
                });

          return (
            <div key={file.name}>
              <Link to={linkTarget} className="flex items-center space-x-3">
                <Icon className="h-5" />
                <span className="flex-1">{file.name}</span>
                <span className="text-gray-8 dark:text-gray-dark-8">
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
