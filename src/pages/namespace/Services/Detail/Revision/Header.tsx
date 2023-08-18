import { pages } from "~/util/router/pages";
import { useServiceRevision } from "~/api/services/query/revision";

const Header = () => {
  const { service, revision } = pages.services.useParams();

  const { data } = useServiceRevision({
    service: service ?? "",
    revision: revision ?? "",
  });

  if (!service) return null;

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col gap-x-7 max-md:space-y-4 md:flex-row md:items-center md:justify-start">
        revisions page: service: {service} - revision: {revision} - {data?.name}
      </div>
    </div>
  );
};

export default Header;
