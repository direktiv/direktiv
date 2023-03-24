import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../../../componentsNext/Dropdown";
import {
  Folder,
  FolderTree,
  FolderUp,
  Github,
  Play,
  PlusCircle,
} from "lucide-react";

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
    <div>
      <div className="space-y-5 border-b border-gray-5 bg-base-200 p-5 dark:border-gray-dark-5">
        <div className="flex flex-col max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between ">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <FolderTree className="h-5" />
            {data?.node?.path}
          </h3>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="primary">
                <PlusCircle /> New <RxChevronDown />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-40">
              <DropdownMenuLabel>Create</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem>
                <Play className="mr-2 h-4 w-4" />
                <span>New Workflow</span>
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Folder className="mr-2 h-4 w-4" />
                <span>New Directory</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
      <div className="flex flex-col space-y-5 p-5 text-sm">
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
            let Icon = Folder;
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
    </div>
  );
};

export default ExplorerPage;
