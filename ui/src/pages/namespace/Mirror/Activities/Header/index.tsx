import EditDialog from "./EditDialog";
import { GitCompare } from "lucide-react";
import { MirrorSchemaType } from "~/api/namespaces/schema";
import SyncDialog from "./SyncDialog";
import { useNamespace } from "~/util/store/namespace";

const Header = ({
  mirror,
  loading,
}: {
  mirror: MirrorSchemaType;
  loading: boolean;
}) => {
  const repoInfo = `${mirror.url} (${mirror.gitRef})`;
  const namespace = useNamespace();

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col gap-x-7 max-md:space-y-4 md:flex-row md:items-center md:justify-start">
        <div className="flex flex-col items-start gap-2">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <GitCompare className="h-5" /> {namespace}
          </h3>
          <div className="text-sm">{repoInfo}</div>
        </div>
        <div className="flex grow justify-end gap-4">
          <EditDialog mirror={mirror} />
          <SyncDialog loading={loading} />
        </div>
      </div>
    </div>
  );
};

export default Header;
