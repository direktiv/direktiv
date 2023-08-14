import { Boxes } from "lucide-react";
import { Card } from "~/design/Card";
import { useServices } from "~/api/services/query/get";

const ServicesListPage = () => {
  const { data } = useServices();

  console.log("🚀", data);

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold">
        <Boxes className="h-5" />
        Services List
      </h3>
      <Card>
        {data?.functions.map((service) => (
          <h1 className="p-2" key={service.info.name}>
            {service.info.name}
          </h1>
        ))}
      </Card>
    </div>
  );
};

export default ServicesListPage;
