import { FC, Fragment } from "react";
import { Link, useParams } from "@tanstack/react-router";

import { FolderTree } from "lucide-react";
import { NewFileDialog } from "./NewFile";
import { analyzePath } from "~/util/router/utils";
import { useNamespace } from "~/util/store/namespace";

const BreadcrumbSegment: FC<{
  absolute: string;
  relative: string;
  namespace: string;
}> = ({ absolute, relative, namespace, ...props }) => (
  <Link
    to="/n/$namespace/explorer/tree/$"
    params={{ namespace, _splat: absolute }}
    className="hover:underline"
    {...props}
  >
    {relative}
  </Link>
);

const ExplorerHeader: FC = () => {
  const namespace = useNamespace();
  const { _splat: path } = useParams({ strict: false });

  const { segments } = analyzePath(path);

  if (!namespace) return null;

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between">
        <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
          <Link
            data-testid="tree-root"
            to="/n/$namespace/explorer/tree/$"
            params={{ namespace, _splat: "" }}
            className="hover:underline"
          >
            <FolderTree className="h-5" />
          </Link>
          <div>
            / &nbsp;
            {segments
              .map((x) => (
                <BreadcrumbSegment
                  key={x.absolute}
                  absolute={x.absolute}
                  relative={x.relative}
                  namespace={namespace}
                />
              ))
              // add / between segments
              .reduce((prev, curr, i) => {
                if (i === 0) return [curr];
                return [
                  ...prev,
                  <Fragment key={`${curr.key}-separator`}> / </Fragment>,
                  curr,
                ];
              }, [] as JSX.Element[])}
          </div>
        </h3>
        <NewFileDialog path={path} />
      </div>
    </div>
  );
};

export default ExplorerHeader;
